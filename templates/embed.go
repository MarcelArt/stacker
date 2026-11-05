// Package templates contains the service templates distributed with Stacker.
package templates

import "embed"

// Files contains the built-in YAML templates.
//
//go:embed *.yml
var Files embed.FS
