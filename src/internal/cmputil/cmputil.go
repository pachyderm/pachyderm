// Package cmputil provides utilities for cmp.Diff.
package cmputil

import (
	"regexp"
	"strings"
	"testing"

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

func WantErr[T ~bool | ~string](t *testing.T, err error, wantErr T) (cont bool) {
	t.Helper()
	var want any = wantErr
	switch w := want.(type) {
	case bool:
		switch {
		case err == nil && w:
			t.Errorf("expected error, but got success")
			return true
		case err != nil && !w:
			t.Errorf("unexpected error: %v", err)
			return false
		case err != nil:
			return false
		case err == nil && !w:
			return true
		}
	case string:
		switch {
		case err == nil && w != "":
			t.Errorf("expected error, but got success")
			return true
		case err != nil && w == "":
			t.Errorf("unexpected error: %v", err)
			return false
		case err != nil:
			if diff := cmp.Diff(w, err.Error(), RegexpStrings()); diff != "" {
				t.Errorf("err: (-want +got)\n%s", diff)
			}
			return false
		case err == nil && w == "":
			return true
		}
	default:
		panic("impossible type passed to WantErr")
	}
	// We enumerate the 4 possible cases in each switch block above, which the compiler doesn't
	// recognize as being complete.  That's fine, if there is a bug, this will panic instead of
	// resulting in a passing test.
	panic("impossible set of circumstances in WantErr")
}
