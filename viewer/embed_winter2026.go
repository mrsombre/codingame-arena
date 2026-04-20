//go:build winter2026

// Package viewer embeds the built frontend bundle for the `arena front` server.
// The //go:build tag matches the game tag passed to `go build`, so a binary
// built with `-tags winter2026` only carries the winter-2026 viewer.
package viewer

import "embed"

//go:embed all:dist/winter-2026
var Assets embed.FS
