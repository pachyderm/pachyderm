package extended

import "testing"

func TestCheckTracesColKey(t *testing.T) {
	var cases = map[string]bool{
		"pipeline0":                   true,
		"project/Pipeline0":           true,
		"project/subproject/pipeline": false, // until hierarchical projects are implemented
		"project 1/pipeline 3":        false,
	}
	for key, expected := range cases {
		if (checkTracesColKey(key) == nil) != expected {
			t.Errorf("checkTracesColKey(%q) is %v", key, !expected)
		}
	}
}
