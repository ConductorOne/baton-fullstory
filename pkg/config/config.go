package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	SCIMBaseURLField = field.StringField(
		"scim-base-url",
		field.WithDisplayName("SCIM Base URL"),
		field.WithDescription("FullStory SCIM base URL (found in Settings > Account Management > SSO)"),
		field.WithRequired(true),
	)

	SCIMTokenField = field.StringField(
		"scim-token",
		field.WithDisplayName("SCIM Token"),
		field.WithDescription("FullStory SCIM bearer token for authentication"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{
		SCIMBaseURLField,
		SCIMTokenField,
	},
	field.WithConnectorDisplayName("Fullstory"),
	field.WithHelpUrl("/docs/baton/fullstory"),
	field.WithIconUrl("/static/app-icons/fullstory.svg"),
)
