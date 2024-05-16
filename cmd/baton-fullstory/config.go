package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	Token string `mapstructure:"token"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.Token == "" {
		return status.Error(codes.InvalidArgument, "api token must be provided, use --help for more information")
	}

	return nil
}

func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("token", "", "FullStory API token. (BATON_FULLSTORY_TOKEN)")
}
