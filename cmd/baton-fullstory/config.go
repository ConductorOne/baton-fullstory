package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	apikey = field.StringField("api-key", field.WithRequired(true), field.WithDescription("FullStory API Key to authenticate with"))
)
