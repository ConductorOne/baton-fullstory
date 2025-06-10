package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIKeyField = field.StringField(
		"api-key",
		field.WithDisplayName("API key"),
		field.WithDescription("FullStory API Key to authenticate with"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{
		APIKeyField,
	},
	field.WithConnectorDisplayName("Fullstory"),
	field.WithHelpUrl("/docs/baton/fullstory"),
	field.WithIconUrl("/static/app-icons/fullstory.svg"),
)
