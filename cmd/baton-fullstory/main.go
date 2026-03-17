package main

import (
	"context"

	cfg "github.com/conductorone/baton-fullstory/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"

	"github.com/conductorone/baton-fullstory/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(ctx,
		"baton-fullstory",
		version,
		cfg.Config,
		connector.New,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilderV2(connector.DefaultCapabilitiesBuilder()),
	)
}
