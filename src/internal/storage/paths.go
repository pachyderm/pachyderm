package storage

import (
	"regexp"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	validRangeRegex = regexp.MustCompile("^[ -~]+$")
	globRegex       = regexp.MustCompile(`[*?[\]{}!()@+^]`)
)

func ValidateFilename(p string) error {
	pBytes := []byte(p)
	if !validRangeRegex.Match(pBytes) {
		return errors.Errorf("path (%v) invalid: only printable ASCII characters allowed", p)
	}
	if globRegex.Match(pBytes) {
		return errors.Errorf("path (%v) invalid: globbing character (%v) not allowed in path", p, globRegex.FindString(p))
	}
	for _, elem := range strings.Split(p, "/") {
		if elem == "." || elem == ".." {
			return errors.Errorf("path (%v) invalid: relative file paths are not allowed", p)
		}
	}
	return nil
}

func GlobLiteralPrefix(glob string) string {
	idx := globRegex.FindStringIndex(glob)
	if idx == nil {
		return glob
	}
	return glob[:idx[0]]
}
