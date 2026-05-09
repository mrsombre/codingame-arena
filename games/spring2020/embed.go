// Package spring2020 bundles game-level metadata (rules, trace format,
// etc.) into the arena binary so `arena game spring2020 rules` and
// `arena game spring2020 trace` ship without a sidecar filesystem path.
package spring2020

import _ "embed"

//go:embed rules.md
var Rules string

//go:embed trace.md
var Trace string
