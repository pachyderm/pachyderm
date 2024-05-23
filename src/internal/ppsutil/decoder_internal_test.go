package ppsutil

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLToJSON(t *testing.T) {
	var cases = map[string]struct {
		input               string // YAML
		expected            string // JSON
		expectYAMLErr       bool
		expectYAMLToJSONErr bool
	}{
		"valid YAML": {
			input: `foo: bar
baz: 0`,
			expected: `{"foo": "bar", "baz": 0}`,
		},
		"invalid YAML": {
			input:         `&foo^baz`,
			expectYAMLErr: true,
		},
		"leading zero": {
			input:               `foo: 01`,
			expectYAMLToJSONErr: true,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var holder yaml.Node
			if err := yaml.Unmarshal([]byte(c.input), &holder); err != nil {
				if c.expectYAMLErr {
					return
				}
				t.Fatal(err)
			}
			if c.expectYAMLErr {
				t.Fatal("got success; expected error")
			}
			got, err := yamlToJSON(holder.Content[0])
			if err != nil {
				if c.expectYAMLToJSONErr {
					return
				}
				t.Fatal(err)
			}
			if c.expectYAMLToJSONErr {
				t.Fatal("got success; expected error")
			}
			var expected any
			d := json.NewDecoder(strings.NewReader(c.expected))
			d.UseNumber()
			if err := d.Decode(&expected); err != nil {
				t.Fatalf("bad expected JSON %s", c.expected)
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("got %v; expected %v", got, expected)
			}
		})
	}
}
