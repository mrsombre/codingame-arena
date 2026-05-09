// Package spring2020 bundles game-level metadata (rules, etc.) into the
// arena binary so `arena game spring2020 rules` ships without a sidecar
// filesystem path.
package spring2020

import _ "embed"

//go:embed rules.md
var Rules string
