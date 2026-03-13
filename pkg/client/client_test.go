package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func newTestServer(t *testing.T, handler http.Handler) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	return NewClient(srv.Client(), srv.URL), srv.Close
}

func scimResponse(w http.ResponseWriter, resp SCIMListResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func TestListUsers_SinglePage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scimResponse(w, SCIMListResponse{
			TotalResults: 2,
			StartIndex:   1,
			ItemsPerPage: 2,
			Resources: []*SCIMUser{
				{ID: "user-1", UserName: "alice", Active: true},
				{ID: "user-2", UserName: "bob", Active: false},
			},
		})
	})

	c, teardown := newTestServer(t, handler)
	defer teardown()

	users, nextStart, err := c.ListUsers(context.Background(), NewPaginationVars(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("got %d users, want 2", len(users))
	}
	if nextStart != 0 {
		t.Errorf("nextStart = %d, want 0 (no more pages)", nextStart)
	}
	if users[0].ID != "user-1" {
		t.Errorf("users[0].ID = %q, want %q", users[0].ID, "user-1")
	}
}

func TestListUsers_MultiPage(t *testing.T) {
	const total = 150

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startIndex := 1
		if v := r.URL.Query().Get("startIndex"); v != "" {
			if idx, err := strconv.Atoi(v); err == nil {
				startIndex = idx
			}
		}

		// Build a page of up to 100 users starting at startIndex.
		var resources []*SCIMUser
		for i := startIndex; i <= total && i < startIndex+defaultCount; i++ {
			resources = append(resources, &SCIMUser{
				ID:       "user-" + strconv.Itoa(i),
				UserName: "user" + strconv.Itoa(i),
			})
		}

		scimResponse(w, SCIMListResponse{
			TotalResults: total,
			StartIndex:   startIndex,
			ItemsPerPage: defaultCount,
			Resources:    resources,
		})
	})

	c, teardown := newTestServer(t, handler)
	defer teardown()

	// First page: expect 100 users and a next start index.
	users, nextStart, err := c.ListUsers(context.Background(), NewPaginationVars(1))
	if err != nil {
		t.Fatalf("page 1: unexpected error: %v", err)
	}
	if len(users) != 100 {
		t.Errorf("page 1: got %d users, want 100", len(users))
	}
	if nextStart != 101 {
		t.Errorf("page 1: nextStart = %d, want 101", nextStart)
	}

	// Second (last) page: expect 50 users and no next start.
	users, nextStart, err = c.ListUsers(context.Background(), NewPaginationVars(nextStart))
	if err != nil {
		t.Fatalf("page 2: unexpected error: %v", err)
	}
	if len(users) != 50 {
		t.Errorf("page 2: got %d users, want 50", len(users))
	}
	if nextStart != 0 {
		t.Errorf("page 2: nextStart = %d, want 0 (last page)", nextStart)
	}
}

func TestListUsers_ExactlyOnePage(t *testing.T) {
	// Total equals exactly one full page: pagination must stop after the first call.
	const total = 100

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resources []*SCIMUser
		for i := 1; i <= total; i++ {
			resources = append(resources, &SCIMUser{ID: "user-" + strconv.Itoa(i)})
		}
		scimResponse(w, SCIMListResponse{
			TotalResults: total,
			StartIndex:   1,
			ItemsPerPage: defaultCount,
			Resources:    resources,
		})
	})

	c, teardown := newTestServer(t, handler)
	defer teardown()

	users, nextStart, err := c.ListUsers(context.Background(), NewPaginationVars(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != total {
		t.Errorf("got %d users, want %d", len(users), total)
	}
	if nextStart != 0 {
		t.Errorf("nextStart = %d, want 0 (all results fit in one page)", nextStart)
	}
}

func TestListUsers_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scimResponse(w, SCIMListResponse{TotalResults: 0})
	})

	c, teardown := newTestServer(t, handler)
	defer teardown()

	users, nextStart, err := c.ListUsers(context.Background(), NewPaginationVars(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("got %d users, want 0", len(users))
	}
	if nextStart != 0 {
		t.Errorf("nextStart = %d, want 0", nextStart)
	}
}

func TestListUsers_HTTPError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})

	c, teardown := newTestServer(t, handler)
	defer teardown()

	_, _, err := c.ListUsers(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

// TestListUsers_TrailingSlashInBaseURL documents the known issue where a
// trailing slash in scimBaseURL produces a double slash in the request path
// (e.g. //Users instead of /Users). This test asserts the correct behavior
// and will fail until the URL construction is fixed.
func TestListUsers_TrailingSlashInBaseURL(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Users" {
			t.Errorf("request path = %q, want %q (double slash from trailing base URL?)", r.URL.Path, "/Users")
		}
		scimResponse(w, SCIMListResponse{TotalResults: 0})
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Trailing slash appended to the base URL.
	c := NewClient(srv.Client(), srv.URL+"/")
	_, _, _ = c.ListUsers(context.Background(), NewPaginationVars(1))
}

func TestListUsers_NilPaginationVarsDefaultsToPage1(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startIndex := r.URL.Query().Get("startIndex")
		if startIndex != "1" {
			t.Errorf("startIndex = %q, want %q", startIndex, "1")
		}
		scimResponse(w, SCIMListResponse{TotalResults: 0})
	})

	c, teardown := newTestServer(t, handler)
	defer teardown()

	_, _, err := c.ListUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
