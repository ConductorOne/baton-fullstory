package main

import (
	cfg "github.com/conductorone/baton-fullstory/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("fullstory", cfg.Config)
}
