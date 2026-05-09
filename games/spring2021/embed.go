// Package spring2021 bundles game-level metadata (rules, trace format,
// etc.) into the arena binary so `arena game spring2021 rules` and
// `arena game spring2021 trace` ship without a sidecar filesystem path.
package spring2021

import _ "embed"

//go:embed rules.md
var Rules string

//go:embed trace.md
var Trace string
