package client

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestUnmarshalTailResponse(t *testing.T) {
	// This was read directly from Loki, but the timestamps were changed to something simpler.
	input := `{"streams":[{"stream":{"app":"a","suite":"pachyderm"},"values":[["1000000000000000000","starting"]]},{"stream":{"app":"a","suite":"pachyderm"},"values":[["1200000000000000000","starting 2"]]},{"stream":{"app":"b","suite":"pachyderm"},"values":[["1300000000000000000","starting"]]}]}`
	var got tailResponse
	if err := json.Unmarshal([]byte(input), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := tailResponse{
		Streams: []tailStream{
			{
				Stream: map[string]string{"app": "a", "suite": "pachyderm"},
				Values: []tailValue{
					{
						Time:     time.Date(2001, 9, 9, 1, 46, 40, 0, time.UTC),
						Messages: []string{"starting"},
					},
				},
			},
			{
				Stream: map[string]string{"app": "a", "suite": "pachyderm"},
				Values: []tailValue{
					{
						Time:     time.Date(2008, 1, 10, 21, 20, 0, 0, time.UTC),
						Messages: []string{"starting 2"},
					},
				},
			},
			{
				Stream: map[string]string{"app": "b", "suite": "pachyderm"},
				Values: []tailValue{
					{
						Time:     time.Date(2011, 3, 13, 7, 6, 40, 0, time.UTC),
						Messages: []string{"starting"},
					},
				},
			},
		},
	}
	require.NoDiff(t, want, got, nil, "json should decode")
}
