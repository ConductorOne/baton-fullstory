package fullstory

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const (
	scimUsersPath = "/Users"
	defaultCount  = 100
)

type Client struct {
	httpClient  *uhttp.BaseHttpClient
	scimBaseURL string
}

func NewClient(client *http.Client, scimBaseURL string) *Client {
	return &Client{
		httpClient:  uhttp.NewBaseHttpClient(client),
		scimBaseURL: scimBaseURL,
	}
}

// PaginationVars holds SCIM pagination state.
// StartIndex is 1-based per the SCIM spec.
type PaginationVars struct {
	StartIndex int
}

func NewPaginationVars(startIndex int) *PaginationVars {
	return &PaginationVars{StartIndex: startIndex}
}

// ListUsers fetches FullStory teammates via the SCIM /Users endpoint.
// Returns the list of users, the next startIndex (0 if no more pages), and any error.
func (c *Client) ListUsers(ctx context.Context, pgVars *PaginationVars) ([]*SCIMUser, int, error) {
	u, err := url.Parse(c.scimBaseURL + scimUsersPath)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing SCIM URL: %w", err)
	}

	startIndex := 1
	if pgVars != nil && pgVars.StartIndex > 0 {
		startIndex = pgVars.StartIndex
	}

	query := url.Values{}
	query.Set("startIndex", strconv.Itoa(startIndex))
	query.Set("count", strconv.Itoa(defaultCount))
	u.RawQuery = query.Encode()

	options := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u, options...)
	if err != nil {
		return nil, 0, fmt.Errorf("creating SCIM request: %w", err)
	}

	var res SCIMListResponse
	resp, err := c.httpClient.Do(req, uhttp.WithJSONResponse(&res))
	if err != nil {
		return nil, 0, fmt.Errorf("SCIM list users: %w", err)
	}
	defer resp.Body.Close()

	// Calculate next page. SCIM startIndex is 1-based.
	nextStart := 0
	if len(res.Resources) > 0 && startIndex+len(res.Resources) <= res.TotalResults {
		nextStart = startIndex + len(res.Resources)
	}

	return res.Resources, nextStart, nil
}
