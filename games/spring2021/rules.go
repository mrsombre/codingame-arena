// Package spring2021 bundles game-level metadata (rules, etc.) into the
// arena binary so `arena game spring2021 rules` ships without a sidecar
// filesystem path.
package spring2021

import _ "embed"

//go:embed rules.md
var Rules string
