package ppstest

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestAPI(t *testing.T, newClient func(t testing.TB) pps.APIClient) {
	tests := []struct {
		Name string
		F    func(t *testing.T, client pps.APIClient)
	}{
		{"NoOp", testNoOp},
		// Add more tests here
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			client := newClient(t)
			test.F(t, client)
		})
	}
}

func testNoOp(t *testing.T, c pps.APIClient) {}
