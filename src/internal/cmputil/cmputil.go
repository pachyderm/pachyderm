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

// WantErr is an assertion helper that checks err against wantErr.  wantErr can be a boolean,
// string, or /string/ which acts like a regular expression against the error message.  WantErr
// returns true if tests should continue to test the non-error case, or false if the test should end
// immediately.
func WantErr[T ~bool | ~string](t *testing.T, err error, wantErr T) (cont bool) {
	t.Helper()
	var want any = wantErr
	switch w := want.(type) {
	case bool:
		switch {
		case err == nil && w:
			t.Errorf("expected error, but got success")
			return false
		case err != nil && !w:
			t.Errorf("unexpected error: %v", err)
		case err != nil:
		case err == nil && !w:
		}
	case string:
		switch {
		case err == nil && w != "":
			t.Errorf("expected error, but got success")
			return false
		case err != nil && w == "":
			t.Errorf("unexpected error: %v", err)
		case err != nil:
			if diff := cmp.Diff(w, err.Error(), RegexpStrings()); diff != "" {
				t.Errorf("err: (-want +got)\n%s", diff)
			}
		case err == nil && w == "":
		}
	default:
		panic("impossible type passed to WantErr")
	}
	return err == nil
}
