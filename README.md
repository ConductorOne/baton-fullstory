![Baton Logo](./docs/images/baton-logo.png)

# `baton-fullstory` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-fullstory.svg)](https://pkg.go.dev/github.com/conductorone/baton-fullstory) ![verify](https://github.com/conductorone/baton-fullstory/actions/workflows/verify.yaml/badge.svg)

`baton-fullstory` is a connector for FullStory built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the FullStory SCIM API to sync data about users.

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Prerequisites

To work with the connector, you need a FullStory SCIM bearer token and your organization's SCIM base URL. Both are found in FullStory under **Settings** > **Account Management** > **SSO**. A user with the **Admin** or **Architect** role must enable SCIM provisioning and generate the token.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-fullstory

BATON_SCIM_BASE_URL=https://app.fullstory.com/scim/v2 BATON_SCIM_TOKEN=your-scim-token baton-fullstory
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out \
  -e BATON_SCIM_BASE_URL=https://app.fullstory.com/scim/v2 \
  -e BATON_SCIM_TOKEN=your-scim-token \
  ghcr.io/conductorone/baton-fullstory:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-fullstory/cmd/baton-fullstory@main

BATON_SCIM_BASE_URL=https://app.fullstory.com/scim/v2 BATON_SCIM_TOKEN=your-scim-token baton-fullstory
baton resources
```

# Data Model

`baton-fullstory` will fetch information about the following FullStory resources:

- Users

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-fullstory` Command Line Usage

```
baton-fullstory

Usage:
  baton-fullstory [flags]
  baton-fullstory [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  config             Get the connector config schema
  health-check       Check the health of a running connector
  help               Help about any command

Flags:
      --client-id string        The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string    The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string             The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                    help for baton-fullstory
      --log-format string       The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string        The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning            This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --scim-base-url string    required: FullStory SCIM base URL (found in Settings > Account Management > SSO) ($BATON_SCIM_BASE_URL)
      --scim-token string       required: FullStory SCIM bearer token for authentication ($BATON_SCIM_TOKEN)
      --skip-full-sync          This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --ticketing               This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                 version for baton-fullstory

Use "baton-fullstory [command] --help" for more information about a command.
```
