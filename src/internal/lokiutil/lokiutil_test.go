package lokiutil

import (
	"encoding/json"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestWrapEntryIfNotJSON(t *testing.T) {
	entries := []streamEntry{
		{Log: "W0627 19:40:45.997742\t 1 warnings.go:70] <msg> duplicate name \"PPS_PIPELINE_NAME\"\n"},
		{Log: "{\"message\":\"started setting up External Pachd GRPC Server\",\"severity\":\"info\"}"},
	}
	for _, entry := range entries {
		t.Run("test wrapEntryIfNotJSON", func(t *testing.T) {
			wrappedEntry, err := wrapEntryIfNotJSON(entry)
			require.NoError(t, err, "entries should be serializable")
			require.NoError(t, json.Unmarshal([]byte(wrappedEntry.Log), &map[string]interface{}{}), "wrappedEntry.Log should be JSON")
		})
	}
}
