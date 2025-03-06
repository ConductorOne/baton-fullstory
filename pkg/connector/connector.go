package connector

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/conductorone/baton-fullstory/pkg/fullstory"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FullStory struct {
	client *fullstory.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (fs *FullStory) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(fs.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (fs *FullStory) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (fs *FullStory) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "FullStory",
		Description: "Connector syncing FullStory users to Baton",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (fs *FullStory) Validate(ctx context.Context) (annotations.Annotations, error) {
	pgVars := fullstory.NewPaginationVars("")
	_, _, err := fs.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "api key is not valid")
	}

	return nil, nil
}

type CustomBasicAuth struct {
	Token string
}

var _ uhttp.AuthCredentials = (*CustomBasicAuth)(nil)

func (c *CustomBasicAuth) GetClient(ctx context.Context, options ...uhttp.Option) (*http.Client, error) {
	httpClient, err := uhttp.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client failed: %w", err)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token, TokenType: "basic"},
	)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, apiKey string) (*FullStory, error) {

	var auth uhttp.AuthCredentials = &uhttp.NoAuth{}
	if apiKey != "" {
		auth = &CustomBasicAuth{Token: apiKey}
	}
	httpClient, err := auth.GetClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	return &FullStory{
		client: fullstory.NewClient(httpClient),
	}, nil
}
