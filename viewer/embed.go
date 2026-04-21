// Package viewer embeds the built frontend bundles for all games.
package viewer

import "embed"

//go:embed all:dist
var Assets embed.FS
