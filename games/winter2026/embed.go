// Package winter2026 bundles game-level metadata (rules, trace format,
// etc.) into the arena binary so `arena game winter2026 rules` and
// `arena game winter2026 trace` ship without a sidecar filesystem path.
package winter2026

import _ "embed"

//go:embed rules.md
var Rules string

//go:embed trace.md
var Trace string
