package ancestry

import (
	"strconv"
	"strings"
)

// Parse parses s for git ancestry references.
// It supports special characters ^ ~, both of which are supported by git.
// Note in git ^ and ~ have different meanings on commits that have multiple
// parent commits. In Pachyderm there's only 1 parent possible so they have
// identical meanings. We support both simply for familiarity sake.
// ParseAncestry returns the base reference and how many ancestors back to go.
// For example:
// foo^ -> foo, 1
// foo^^ -> foo, 2
// foo^3 -> foo 3
// (all examples apply with ~ in place of ^ as well
func Parse(s string) (string, int) {
	sepIndex := strings.IndexAny(s, "^~")
	if sepIndex == -1 {
		return s, 0
	}

	// Find the separator, which is either "^" or "~"
	sep := s[sepIndex]
	strAfterSep := s[sepIndex+1:]

	// Try convert the string after the separator to an int.
	intAfterSep, err := strconv.Atoi(strAfterSep)
	// If it works, return
	if err == nil {
		return s[:sepIndex], intAfterSep
	}

	// Otherwise, we check if there's a sequence of separators, as in
	// "master^^^^" or "master~~~~"
	for i := sepIndex + 1; i < len(s); i++ {
		if s[i] != sep {
			// If we find a character that's not the separator, as in
			// "master~whatever", then we return.
			return s, 0
		}
	}

	// Here we've confirmed that the commit ID ends with a sequence of
	// (the same) separators and therefore uses the correct ancestry
	// syntax.
	return s[:sepIndex], len(s) - sepIndex
}
