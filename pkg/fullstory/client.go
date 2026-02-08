package fullstory

import (
	"context"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const (
	DefaultBaseURL = "https://api.fullstory.com"

	V2Endpoint    = "/v2"
	UsersEndpoint = "/users"
)

type Client struct {
	httpClient *uhttp.BaseHttpClient
	baseURL    string
}

func NewClient(client *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		httpClient: uhttp.NewBaseHttpClient(client),
		baseURL:    baseURL,
	}
}

func (c *Client) prepareURL(path string) (*url.URL, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	p, err := url.JoinPath(V2Endpoint, path)
	if err != nil {
		return nil, err
	}

	base.Path = p
	return base, nil
}

type PaginationVars struct {
	NextPageToken string `json:"nextPageToken"`
}

func NewPaginationVars(nextPageToken string) *PaginationVars {
	return &PaginationVars{
		NextPageToken: nextPageToken,
	}
}

type ListResponse[T any] struct {
	Results  []T    `json:"results"`
	NextPage string `json:"next_page_token"`
}

func (c *Client) ListUsers(ctx context.Context, pgVars *PaginationVars) ([]User, string, error) {
	options := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
	}

	u, err := c.prepareURL(UsersEndpoint)
	if err != nil {
		return nil, "", err
	}

	req, err := c.httpClient.NewRequest(ctx, http.MethodGet, u, options...)
	if err != nil {
		return nil, "", err
	}

	if pgVars != nil && pgVars.NextPageToken != "" {
		query := url.Values{}
		query.Set("page_token", pgVars.NextPageToken)
		req.URL.RawQuery = query.Encode()
	}

	var res ListResponse[User]
	resp, err := c.httpClient.Do(req, uhttp.WithJSONResponse(&res))
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()

	return res.Results, res.NextPage, nil
}
