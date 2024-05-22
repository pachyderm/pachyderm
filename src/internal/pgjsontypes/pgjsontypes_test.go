package pgjsontypes

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNonStringData(t *testing.T) {
	var got StringMap
	if err := got.Scan([]byte(`{"good": "value", "bad": [1, 2, 3]}`)); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"good": "value",
		"bad":  "[1, 2, 3]",
	}
	if diff := cmp.Diff(want, got.Data); diff != "" {
		t.Errorf("Scan (-want +got):\n%s", diff)
	}
}

func TestStringMapRoundTrip(t *testing.T) {
	testData := []struct {
		name string
		data map[string]string
	}{
		{
			name: "nil",
			data: nil,
		},
		{
			name: "empty",
			data: make(map[string]string),
		},
		{
			name: "valid",
			data: map[string]string{"key": "value", "key2": "value2"},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			js, err := (&StringMap{Data: test.data}).Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}
			var got StringMap
			if err := got.Scan(js); err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if diff := cmp.Diff(test.data, got.Data, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("RoundTrip (-want +got):\n%s", diff)
			}
		})
	}
}
