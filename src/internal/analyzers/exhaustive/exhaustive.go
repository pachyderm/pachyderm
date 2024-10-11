// Package exhaustive provides an analyzer from github.com/nishanths/exhaustive.
package exhaustive

import "github.com/nishanths/exhaustive"

var Analyzer = exhaustive.Analyzer

func init() {
	if err := Analyzer.Flags.Lookup(exhaustive.DefaultSignifiesExhaustiveFlag).Value.Set("true"); err != nil {
		panic("exhaustive: failed to set default-signifies-exhaustive")
	}
	if err := Analyzer.Flags.Lookup(exhaustive.ExplicitExhaustiveMapFlag).Value.Set("true"); err != nil {
		panic("exhaustive: failed to set explicit-exhaustive-map")
	}
	if err := Analyzer.Flags.Lookup(exhaustive.ExplicitExhaustiveSwitchFlag).Value.Set("true"); err != nil {
		panic("exhaustive: failed to set explicit-exhaustive-switch")
	}
}
