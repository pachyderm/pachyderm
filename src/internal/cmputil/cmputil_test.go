package cmputil

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRegexpStrings(t *testing.T) {
	testData := []struct {
		name     string
		a, b     []string
		wantDiff bool
	}{
		{
			name: "regexp_b",
			a:    []string{"foo", "bar", "baz"},
			b:    []string{"/^foo$/", "/^ba/", "/^ba/"},
		},
		{
			name:     "regexp_b_mismatch",
			a:        []string{"foo", "bar", "baz"},
			b:        []string{"/^food$/", "/^ba/", "/^ba/"},
			wantDiff: true,
		},
		{
			name: "regexp_a",
			a:    []string{"/^foo$/", "/^ba/", "/^ba/"},
			b:    []string{"foo", "bar", "baz"},
		},
		{
			name: "literal_match",
			a:    []string{"a"},
			b:    []string{"a"},
		},
		{
			name:     "literal_mismatch",
			a:        []string{"a"},
			b:        []string{"b"},
			wantDiff: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			diff := cmp.Diff(test.a, test.b, RegexpStrings())
			if !test.wantDiff && diff != "" {
				t.Errorf("unexpected diff:\n%s", diff)
			} else if test.wantDiff && diff == "" {
				t.Error("diff empty, but wanted one")
			}
		})
	}
}

func TestWantErr(t *testing.T) {
	WantErr(t, nil, false)
	WantErr(t, nil, "")
	WantErr(t, errors.New("blah"), true)
	WantErr(t, errors.New("blah"), "/blah/")
}
