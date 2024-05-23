// Package jsonschema bundles the generated JSON schemas.
package jsonschema

import "embed"

//go:embed */*.json
var FS embed.FS
