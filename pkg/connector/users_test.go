package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/conductorone/baton-fullstory/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

// --- createResource unit tests ---

func TestCreateResource_ActiveUser(t *testing.T) {
	b := newUserBuilder(nil)
	user := &client.SCIMUser{
		ID:          "scim-123",
		UserName:    "alice",
		DisplayName: "Alice Smith",
		Emails:      []*client.SCIMEmail{{Value: "alice@example.com", Primary: true}},
		Active:      true,
	}

	res, err := b.createResource(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.GetId().GetResource() != "scim-123" {
		t.Errorf("resource ID = %q, want %q", res.GetId().GetResource(), "scim-123")
	}
	if res.GetId().GetResourceType() != "user" {
		t.Errorf("resource type = %q, want %q", res.GetId().GetResourceType(), "user")
	}
	if res.GetDisplayName() != "Alice Smith" {
		t.Errorf("display name = %q, want %q", res.GetDisplayName(), "Alice Smith")
	}

	trait, err := rs.GetUserTrait(res)
	if err != nil {
		t.Fatalf("failed to get user trait: %v", err)
	}
	if trait.GetLogin() != "alice" {
		t.Errorf("login = %q, want %q", trait.GetLogin(), "alice")
	}
	if trait.GetStatus().GetStatus() != v2.UserTrait_Status_STATUS_ENABLED {
		t.Errorf("status = %v, want STATUS_ENABLED", trait.GetStatus().GetStatus())
	}
	if len(trait.GetEmails()) == 0 || trait.GetEmails()[0].GetAddress() != "alice@example.com" {
		t.Errorf("expected primary email alice@example.com, got %v", trait.GetEmails())
	}
}

func TestCreateResource_InactiveUser(t *testing.T) {
	b := newUserBuilder(nil)
	user := &client.SCIMUser{
		ID:       "scim-456",
		UserName: "bob",
		Active:   false,
	}

	res, err := b.createResource(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trait, err := rs.GetUserTrait(res)
	if err != nil {
		t.Fatalf("failed to get user trait: %v", err)
	}
	if trait.GetStatus().GetStatus() != v2.UserTrait_Status_STATUS_DISABLED {
		t.Errorf("status = %v, want STATUS_DISABLED", trait.GetStatus().GetStatus())
	}
}

func TestCreateResource_UsesSCIMIDAsResourceID(t *testing.T) {
	b := newUserBuilder(nil)
	user := &client.SCIMUser{
		ID:       "stable-scim-id",
		UserName: "charlie",
		Active:   true,
	}

	res, err := b.createResource(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The resource ID must be the SCIM id, not the username or any derived value,
	// to ensure stability across syncs.
	if res.GetId().GetResource() != user.ID {
		t.Errorf("resource ID = %q, want SCIM ID %q", res.GetId().GetResource(), user.ID)
	}
}

func TestCreateResource_ProfileFields(t *testing.T) {
	b := newUserBuilder(nil)
	user := &client.SCIMUser{
		ID:       "scim-789",
		UserName: "dave",
		Name:     client.SCIMName{GivenName: "Dave", FamilyName: "Jones"},
		Active:   true,
	}

	res, err := b.createResource(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trait, err := rs.GetUserTrait(res)
	if err != nil {
		t.Fatalf("failed to get user trait: %v", err)
	}

	profile := trait.GetProfile()
	if first, ok := rs.GetProfileStringValue(profile, "first_name"); !ok || first != "Dave" {
		t.Errorf("first_name = %q, want %q", first, "Dave")
	}
	if last, ok := rs.GetProfileStringValue(profile, "last_name"); !ok || last != "Jones" {
		t.Errorf("last_name = %q, want %q", last, "Jones")
	}
	if scimID, ok := rs.GetProfileStringValue(profile, "scim_id"); !ok || scimID != "scim-789" {
		t.Errorf("scim_id = %q, want %q", scimID, "scim-789")
	}
}

// TestCreateResource_NoEmailFallsBackToUserName documents the behavior when
// a SCIM user has no emails array. GetEmail() falls back to UserName as the
// email value. Verify FullStory's SCIM API always populates emails for active
// users to determine if this fallback is ever reached in practice.
func TestCreateResource_NoEmailFallsBackToUserName(t *testing.T) {
	b := newUserBuilder(nil)
	user := &client.SCIMUser{
		ID:       "scim-000",
		UserName: "nomail",
		Active:   true,
		// Emails intentionally omitted.
	}

	res, err := b.createResource(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trait, err := rs.GetUserTrait(res)
	if err != nil {
		t.Fatalf("failed to get user trait: %v", err)
	}

	if len(trait.GetEmails()) == 0 {
		t.Fatal("expected email to be set via UserName fallback")
	}
	if trait.GetEmails()[0].GetAddress() != "nomail" {
		t.Errorf("fallback email address = %q, want %q", trait.GetEmails()[0].GetAddress(), "nomail")
	}
}

// --- List integration tests (uses httptest server) ---

func newSCIMServer(t *testing.T, handler http.Handler) (*client.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	return client.NewClient(srv.Client(), srv.URL), srv.Close
}

func encodeSCIM(w http.ResponseWriter, resp client.SCIMListResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func syncAttrs(pageToken string) rs.SyncOpAttrs {
	return rs.SyncOpAttrs{PageToken: pagination.Token{Token: pageToken}}
}

func TestList_FirstPage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeSCIM(w, client.SCIMListResponse{
			TotalResults: 3,
			StartIndex:   1,
			ItemsPerPage: 100,
			Resources: []*client.SCIMUser{
				{ID: "u1", UserName: "alice", Active: true},
				{ID: "u2", UserName: "bob", Active: false},
				{ID: "u3", UserName: "carol", Active: true},
			},
		})
	})

	c, teardown := newSCIMServer(t, handler)
	defer teardown()

	b := newUserBuilder(c)
	resources, results, err := b.List(context.Background(), nil, syncAttrs(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 3 {
		t.Errorf("got %d resources, want 3", len(resources))
	}
	// Single page: next token should be empty (pagination done).
	if results.NextPageToken != "" {
		t.Errorf("NextPageToken = %q, want empty (no more pages)", results.NextPageToken)
	}
}

func TestList_PaginationTwoPages(t *testing.T) {
	const total = 150

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startIndex := 1
		if v := r.URL.Query().Get("startIndex"); v != "" {
			if idx, err := strconv.Atoi(v); err == nil {
				startIndex = idx
			}
		}

		var resources []*client.SCIMUser
		end := startIndex + 100
		if end > total+1 {
			end = total + 1
		}
		for i := startIndex; i < end; i++ {
			resources = append(resources, &client.SCIMUser{
				ID:       "u" + strconv.Itoa(i),
				UserName: "user" + strconv.Itoa(i),
				Active:   true,
			})
		}

		encodeSCIM(w, client.SCIMListResponse{
			TotalResults: total,
			StartIndex:   startIndex,
			ItemsPerPage: 100,
			Resources:    resources,
		})
	})

	c, teardown := newSCIMServer(t, handler)
	defer teardown()

	b := newUserBuilder(c)

	// First page.
	resources, results, err := b.List(context.Background(), nil, syncAttrs(""))
	if err != nil {
		t.Fatalf("page 1: unexpected error: %v", err)
	}
	if len(resources) != 100 {
		t.Errorf("page 1: got %d resources, want 100", len(resources))
	}
	if results.NextPageToken == "" {
		t.Fatal("page 1: expected non-empty NextPageToken")
	}

	// Second (last) page using the token from the first.
	resources, results, err = b.List(context.Background(), nil, syncAttrs(results.NextPageToken))
	if err != nil {
		t.Fatalf("page 2: unexpected error: %v", err)
	}
	if len(resources) != 50 {
		t.Errorf("page 2: got %d resources, want 50", len(resources))
	}
	if results.NextPageToken != "" {
		t.Errorf("page 2: NextPageToken = %q, want empty (last page)", results.NextPageToken)
	}
}

func TestList_ResourceIDsAreStable(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeSCIM(w, client.SCIMListResponse{
			TotalResults: 1,
			StartIndex:   1,
			ItemsPerPage: 100,
			Resources: []*client.SCIMUser{
				{ID: "scim-stable-id", UserName: "alice@example.com", Active: true},
			},
		})
	})

	c, teardown := newSCIMServer(t, handler)
	defer teardown()

	b := newUserBuilder(c)
	resources, _, err := b.List(context.Background(), nil, syncAttrs(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(resources))
	}

	if resources[0].GetId().GetResource() != "scim-stable-id" {
		t.Errorf("resource ID = %q, want %q", resources[0].GetId().GetResource(), "scim-stable-id")
	}
}

func TestList_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})

	c, teardown := newSCIMServer(t, handler)
	defer teardown()

	b := newUserBuilder(c)
	_, _, err := b.List(context.Background(), nil, syncAttrs(""))
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}
