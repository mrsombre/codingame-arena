// Package winter2026 bundles game-level metadata (rules, etc.) into the
// arena binary so `arena game winter2026 rules` ships without a sidecar
// filesystem path.
package winter2026

import _ "embed"

//go:embed rules.md
var Rules string
