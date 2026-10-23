// Package schema provides versioned publication schemas and compatibility fixtures.
package schema

import "embed"

// FS holds the publication JSON Schemas and the fixtures they are tested against.
//
//go:embed *.json fixtures/*.json
var FS embed.FS
