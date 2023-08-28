package ppsutil_test

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
)

func TestSpecReader(t *testing.T) {
	var cases = map[string]struct {
		input             string // YAML
		disableValidation bool
		expected          []string // list of JSON specs
		expectErr         bool
	}{
		"empty input is okay": {},
		"invalid JSON is bad": {
			input:     `"dkjfskjf`,
			expectErr: true,
		},
		"valid JSON is valid": {
			input:    `{"pipeline": {"name": "test"}}`,
			expected: []string{`{"pipeline": {"name": "test"}}`},
		},
		"valid spec is valid": {
			input:    `{"pipeline": {"name": "test"}, "datumTries": 1, "resourceRequests":{"disk":"256Mi", "cpu": .5}, "autoscaling":true}`,
			expected: []string{`{"pipeline": {"name": "test"},"resourceRequests":{"disk":"256Mi", "cpu": 0.5},"datumTries": 1, "autoscaling":true}`},
		},
		"invalid spec is invalid": {
			input:     `{"not_aField": 2}`,
			expectErr: true,
		},
		"multiple specs work": {
			input: `---
pipeline:
  name: test
...
---
{"pipeline": {"name": "test2"}}
`,
			expected: []string{`{"pipeline": {"name": "test"}}`, `{"pipeline": {"name": "test2"}}`},
		},
		"lists work": {
			input: `
- {"pipeline": {"name": "test"}}
- {"pipeline": {"name": "test2"}}`,
			expected: []string{`{"pipeline": {"name": "test"}}`, `{"pipeline": {"name": "test2"}}`},
		},
		"null works": {
			input: `
                                  pipeline:
                                    name: test
                                  resourceRequests: null`,
			expected: []string{`{"pipeline": {"name": "test"},"resourceRequests": null}`},
		},
		"sequences work": {
			input: `
                                 pipeline:
                                   name: test
                                 transform:
                                   cmd: ["foo", "bar"]`,
			expected: []string{`{"pipeline": {"name": "test"}, "transform": {"cmd": ["foo", "bar"]}}`},
		},
		// TODO(INT-1006): This test should be removed when INT-1006 is implemented.
		"disabling validation works": {
			input:             `{"input": {"pfs": {"project": "foo", "repo": "repo", "glob": "/"}}}`,
			disableValidation: true,
			expected:          []string{`{"input": {"pfs": {"project": "foo", "repo": "repo", "glob": "/"}}}`},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				r   = ppsutil.NewSpecReader(strings.NewReader(c.input))
				got []string
				i   int
			)
			if c.disableValidation {
				r = r.DisableValidation()
			}
			for {
				g, err := r.Next()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					if !c.expectErr {
						t.Errorf("item %d: %v", i, err)
					}
					return
				}
				got = append(got, g)
				i++
			}
			if c.expectErr {
				t.Fatalf("got success; expected error")
			}
			if len(got) != len(c.expected) {
				t.Fatalf("got len(%v) = %d; expected %d", got, len(got), len(c.expected))
			}
			for i := 0; i < len(got); i++ {
				var g, e any
				d := json.NewDecoder(strings.NewReader(got[i]))
				d.UseNumber()
				if err := d.Decode(&g); err != nil {
					t.Fatalf("item %d: could not unmarshal %s: %v", i, got[i], err)
				}
				d = json.NewDecoder(strings.NewReader(c.expected[i]))
				d.UseNumber()
				if err := d.Decode(&e); err != nil {
					t.Fatalf("item %d: could not unmarshal %s: %v", i, c.expected[i], err)
				}
				if !reflect.DeepEqual(g, e) {
					t.Fatalf("item %d: got %v; expected %v", i, got[i], c.expected[i])
				}
			}
		})
	}
}
