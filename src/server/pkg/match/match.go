package match

import (
	"fmt"
	"path"
	"strings"
)

// Match is the same as path.Match except that it supports the double
// wildcard pattern "**".  Specifally we only support one "**".  Supporting
// multiple "**" in the same pattern significantly complicates the algorithm.
//
// This link explains "**": https://stackoverflow.com/a/28199633/1563568
func Match(pattern, name string) (bool, error) {
	// Use the standard implementation if the pattern does not contain "**"
	if !strings.Contains(pattern, "**") {
		return path.Match(pattern, name)
	}

	// The algorithm is simple: we divide the pattern into two sub-patterns:
	// one to the left of "**" and the other to the right of "**".  We then
	// use the left sub-pattern to match the name from the left, and use the
	// right sub-pattern to match the name from the right.  The pattern
	// matches if the following are true:
	//
	// * Both sub-patterns match.
	// * The sub-patterns do not match overlapping parts.

	subPatterns := strings.Split(pattern, "**")
	if len(subPatterns) != 2 {
		return false, path.ErrBadPattern
	}

	// This is a test for the case where the pattern has more than 2
	// consecutive asterisks.
	if strings.HasPrefix(subPatterns[1], "*") {
		return false, path.ErrBadPattern
	}

	leftPats := rmEmptyStrs(strings.Split(subPatterns[0], "/"))
	rightPats := rmEmptyStrs(strings.Split(subPatterns[1], "/"))

	components := rmEmptyStrs(strings.Split(name, "/"))
	// Check overlaps
	if len(components) < len(leftPats)+len(rightPats) {
		return false, nil
	}

	return matchComponents(leftPats, components[:len(leftPats)]) && matchComponents(rightPats, components[len(components)-len(rightPats):]), nil
}

// matchComponents matches patterns against components of a path
func matchComponents(pats, comps []string) bool {
	fmt.Printf("matching %v with %v\n", pats, comps)
	if len(pats) != len(comps) {
		return false
	}
	for i, pat := range pats {
		if pat != "*" && pat != comps[i] {
			return false
		}
	}
	return true
}

// rmEmptyStrs returns a slice of strings that's equivalent to the given
// slice but with empty strings removed.
func rmEmptyStrs(strs []string) []string {
	var ret []string
	for _, s := range strs {
		if s != "" {
			ret = append(ret, s)
		}
	}
	return ret
}
