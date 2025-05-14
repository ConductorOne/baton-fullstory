package main

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/conductorone/baton-fullstory/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	"github.com/conductorone/baton-fullstory/pkg/connector"
	configschema "github.com/conductorone/baton-sdk/pkg/config"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := configschema.DefineConfiguration(ctx, "baton-fullstory", getConnector, cfg.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, fsc *cfg.Fullstory) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	apiKey := fsc.ApiKey

	cb, err := connector.New(ctx, apiKey)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return c, nil
}
