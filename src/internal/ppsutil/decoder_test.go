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
		input     string   // YAML
		expected  []string // list of JSON specs
		expectErr bool
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
			input: `{"pipeline": {"name": "test"}}
---
{"pipeline": {"name": "test2"}}`,
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
	}
cases:
	for n, c := range cases {
		var (
			r   = ppsutil.NewSpecReader(strings.NewReader(c.input))
			got []string
			i   int
		)
		for {
			g, err := r.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				if !c.expectErr {
					t.Errorf("%s[%d]: %v", n, i, err)
				}
				continue cases
			}
			got = append(got, g)
			i++
		}
		if err := r.Error(); err != nil {
			if !c.expectErr {
				t.Errorf("%s: %v", n, err)
			}
			continue
		}
		if c.expectErr {
			t.Errorf("%s: error expected; got %v", n, got)
			continue
		}
		if len(got) != len(c.expected) {
			t.Errorf("%s: expected %d results; got %d", n, len(c.expected), len(got))
			continue
		}
		for i := 0; i < len(got); i++ {
			var g, e any
			d := json.NewDecoder(strings.NewReader(got[i]))
			d.UseNumber()
			if err := d.Decode(&g); err != nil {
				t.Errorf("%s[%d]: could not unmarshal %s", n, i, got[i])
				continue
			}
			d = json.NewDecoder(strings.NewReader(c.expected[i]))
			d.UseNumber()
			if err := d.Decode(&e); err != nil {
				t.Errorf("%s[%d]: could not unmarshal %s", n, i, c.expected[i])
				continue
			}
			if !reflect.DeepEqual(g, e) {
				t.Errorf("%s[%d]: expected %v; got %v", n, i, c.expected[i], got[i])
				continue
			}
		}
	}
}
