// Package winter2026 bundles game-level metadata (rules, trace format,
// etc.) into the arena binary so `arena game rules winter2026` and
// `arena game trace winter2026` ship without a sidecar filesystem path.
package winter2026

import _ "embed"

//go:embed rules.md
var Rules string

//go:embed trace.md
var Trace string
