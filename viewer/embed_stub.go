//go:build !winter2026

// Stub so the package compiles when no game tag is set. The `arena front`
// command is unreachable in this configuration — main.go exits early when
// arena.Factory is nil — but the package must still build.
package viewer

import "embed"

//go:embed embed_stub.go
var Assets embed.FS
