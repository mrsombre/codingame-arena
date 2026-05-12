// Package spring2026 bundles game-level metadata (rules, etc.) into the
// arena binary so `arena game rules spring2026` ships without a sidecar
// filesystem path.
package spring2026

import _ "embed"

//go:embed rules.md
var Rules string
