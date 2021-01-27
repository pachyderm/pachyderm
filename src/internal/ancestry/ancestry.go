package ancestry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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
		if sep == '.' {
			return s[:sepIndex], -1 * intAfterSep, nil
		}
		return s[:sepIndex], intAfterSep, nil
	}

	// Otherwise, we check if there's a sequence of separators, as in
	// "master^^^^" or "master~~~~"
	for i := sepIndex + 1; i < len(s); i++ {
		if s[i] != sep {
			// If we find a character that's not the separator, as in
			// "master~whatever", then we return an error
			return "", 0, errors.Errorf("invalid ancestry syntax %q, cannot mix %c and %c characters", s, sep, s[i])
		}
	}

	// Here we've confirmed that the commit ID ends with a sequence of
	// (the same) separators and therefore uses the correct ancestry
	// syntax.
	if sep == '.' {
		return s[:sepIndex], -1*len(s) - sepIndex, nil
	}
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

var (
	valid              = regexp.MustCompile("^[a-zA-Z0-9_-]+$") // Matches a valid name
	invalid            = regexp.MustCompile("[^a-zA-Z0-9_-]")   // matches an invalid character
	invalidNameErrorRe = regexp.MustCompile(`name \(.+\) invalid: only alphanumeric characters, underscores, and dashes are allowed`)
)

// ValidateName validates a name to make sure that it can be used unambiguously
// with Ancestry syntax.
func ValidateName(name string) error {
	if !valid.MatchString(name) {
		return errors.Errorf("name (%v) invalid: only alphanumeric characters, underscores, and dashes are allowed", name)
	}
	return nil
}

// SanitizeName forces a name to pass ValidateName, by replacing offending
// characters with _s
func SanitizeName(name string) string {
	return invalid.ReplaceAllString(name, "_")
}

// IsInvalidNameError returns true if err is due to an invalid name
func IsInvalidNameError(err error) bool {
	if err == nil {
		return false
	}
	return invalidNameErrorRe.MatchString(err.Error())
}
