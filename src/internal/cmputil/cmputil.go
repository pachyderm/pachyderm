// Package cmputil provides utilities for cmp.Diff.
package cmputil

import (
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// RegexpStrings treats strings that start and end with / as regular expressions, i.e. "/foo/" would
// match "foobar".
func RegexpStrings() cmp.Option {
	return cmp.Comparer(func(a, b string) bool {
		if strings.HasPrefix(a, "/") && strings.HasSuffix(a, "/") && len(a) > 2 {
			rx := regexp.MustCompile(a[1 : len(a)-1])
			return rx.MatchString(b)
		} else if strings.HasPrefix(b, "/") && strings.HasSuffix(b, "/") && len(b) > 2 {
			rx := regexp.MustCompile(b[1 : len(b)-1])
			return rx.MatchString(a)
		}
		return a == b
	})
}
