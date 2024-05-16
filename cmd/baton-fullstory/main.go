package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/conductorone/baton-fullstory/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	cfg := &config{}
	cmd, err := cli.NewCmd(ctx, "baton-fullstory", cfg, validateConfig, getConnector)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version
	cmdFlags(cmd)

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
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

func getConnector(ctx context.Context, cfg *config) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	var auth uhttp.AuthCredentials = &uhttp.NoAuth{}
	if cfg.Token != "" {
		auth = &CustomBasicAuth{Token: cfg.Token}
	}

	cb, err := connector.New(ctx, auth)
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
