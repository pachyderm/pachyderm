package ancestry

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses s for git ancestry references.
// It supports special characters ^ ~, both of which are supported by git.
// Note in git ^ and ~ have different meanings on commits that have multiple
// parent commits. In Pachyderm there's only 1 parent possible so they have
// identical meanings. We support both simply for familiarity sake.
// In addition we support referencing the beginning of the branch with the . character.
// ParseAncestry returns the base reference and how many ancestors back to go.
// For example:
// foo^ -> foo, 1
// foo^^ -> foo, 2
// foo^3 -> foo, 3
// foo.1 -> foo, -1
// foo.3 -> foo, -3
// (all examples apply with ~ in place of ^ as well
func Parse(s string) (string, int, error) {
	sepIndex := strings.IndexAny(s, "^~.")
	if sepIndex == -1 {
		return s, 0, nil
	}

	// Find the separator, which is either "^" or "~"
	sep := s[sepIndex]
	strAfterSep := s[sepIndex+1:]

	// Try convert the string after the separator to an int.
	intAfterSep, err := strconv.Atoi(strAfterSep)
	// If it works, return
	if err == nil {
		return s[:sepIndex], intAfterSep, nil
	}

	// Otherwise, we check if there's a sequence of separators, as in
	// "master^^^^" or "master~~~~"
	for i := sepIndex + 1; i < len(s); i++ {
		if s[i] != sep {
			// If we find a character that's not the separator, as in
			// "master~whatever", then we return an error
			return "", 0, fmt.Errorf("invalid ancestry syntax %q, cannot mix %c and %c characters", sep, s[i])
		}
	}

	// Here we've confirmed that the commit ID ends with a sequence of
	// (the same) separators and therefore uses the correct ancestry
	// syntax.
	return s[:sepIndex], len(s) - sepIndex, nil
}

// Add adds an ancestry reference to the given string.
func Add(s string, ancestors int) string {
	if ancestors > 0 {
		return fmt.Sprintf("%s~%d", s, ancestors)
	} else if ancestors < 0 {
		return fmt.Sprintf("%s.%d", s, ancestors*-1)
	}
	return s
}
