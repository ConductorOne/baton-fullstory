package connector

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/conductorone/baton-fullstory/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"

	cfg "github.com/conductorone/baton-fullstory/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"golang.org/x/oauth2"
)

type FullStory struct {
	client *client.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (fs *FullStory) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(fs.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (fs *FullStory) Asset(_ context.Context, _ *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (fs *FullStory) Metadata(_ context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "FullStory",
		Description: "Connector for FullStory that allows you to manage teammates",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (fs *FullStory) Validate(ctx context.Context) (annotations.Annotations, error) {
	pgVars := client.NewPaginationVars(1)
	_, _, err := fs.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, fmt.Errorf("baton-fullstory: validation failed with error: %w", err)
	}

	return nil, nil
}

type defaultCapabilitiesBuilder struct{}

// DefaultCapabilitiesBuilder returns all resource types unconditionally so that
// the generated capabilities are always complete regardless of connector configuration.
func DefaultCapabilitiesBuilder() connectorbuilder.ConnectorBuilderV2 {
	return &defaultCapabilitiesBuilder{}
}

func (d *defaultCapabilitiesBuilder) Metadata(_ context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "FullStory",
		Description: "Connector for FullStory that allows you to manage teammates",
	}, nil
}

func (d *defaultCapabilitiesBuilder) Validate(_ context.Context) (annotations.Annotations, error) {
	return nil, nil
}

func (d *defaultCapabilitiesBuilder) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(nil),
	}
}

// SCIMBearerAuth implements Bearer token authentication for SCIM.
type SCIMBearerAuth struct {
	Token string
}

var _ uhttp.AuthCredentials = (*SCIMBearerAuth)(nil)

func (c *SCIMBearerAuth) GetClient(ctx context.Context, options ...uhttp.Option) (*http.Client, error) {
	httpClient, err := uhttp.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("baton-fullstory: creating HTTP client failed: %w", err)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token, TokenType: "Bearer"},
	)
	httpClient = oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

func New(ctx context.Context, fsc *cfg.Fullstory, _ *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	auth := &SCIMBearerAuth{Token: fsc.ScimToken}
	httpClient, err := auth.GetClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, nil, err
	}

	return &FullStory{
		client: client.NewClient(httpClient, fsc.ScimBaseUrl),
	}, nil, nil
}
