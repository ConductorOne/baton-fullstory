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

func (fs *FullStory) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(fs.client),
	}
}

func (fs *FullStory) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

func (fs *FullStory) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "FullStory",
		Description: "Connector syncing FullStory teammates to Baton via SCIM",
	}, nil
}

func (fs *FullStory) Validate(ctx context.Context) (annotations.Annotations, error) {
	pgVars := fullstory.NewPaginationVars(1)
	_, _, err := fs.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "SCIM credentials are not valid")
	}

	return nil, nil
}

// SCIMBearerAuth implements Bearer token authentication for SCIM.
type SCIMBearerAuth struct {
	Token string
}

var _ uhttp.AuthCredentials = (*SCIMBearerAuth)(nil)

func (c *SCIMBearerAuth) GetClient(ctx context.Context, options ...uhttp.Option) (*http.Client, error) {
	httpClient, err := uhttp.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client failed: %w", err)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token, TokenType: "Bearer"},
	)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

func New(ctx context.Context, scimBaseURL string, scimToken string) (*FullStory, error) {
	var auth uhttp.AuthCredentials = &uhttp.NoAuth{}
	if scimToken != "" {
		auth = &SCIMBearerAuth{Token: scimToken}
	}
	httpClient, err := auth.GetClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	return &FullStory{
		client: fullstory.NewClient(httpClient, scimBaseURL),
	}, nil
}
