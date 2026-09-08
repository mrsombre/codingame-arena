// Package summer2026 bundles game-level metadata (rules, etc.) into the
// arena binary so `arena game rules summer2026` ships without a sidecar
// filesystem path.
package summer2026

import _ "embed"

//go:embed rules.md
var Rules string

//go:embed trace.md
var Trace string
