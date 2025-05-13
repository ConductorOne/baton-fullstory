package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIKeyField = field.StringField("api-key", field.WithRequired(true), field.WithDescription("FullStory API Key to authenticate with"))
)

//go:generate go run ./gen
var Config = field.NewConfiguration([]field.SchemaField{
	APIKeyField,
})
