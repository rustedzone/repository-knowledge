package repositoryknowledge

import (
	"embed"
	"strings"
)

// Content contains every toolkit-managed asset installed into consuming repositories.
//
//go:embed VERSION policy/*.json schemas/*.json templates/* skills/repository-knowledge/*.md skills/repository-knowledge/references/*.md skills/repository-knowledge/scripts/*
var Content embed.FS

// Version returns the semantic version embedded into this binary.
func Version() string {
	value, err := Content.ReadFile("VERSION")
	if err != nil {
		return "development"
	}
	return strings.TrimSpace(string(value))
}
