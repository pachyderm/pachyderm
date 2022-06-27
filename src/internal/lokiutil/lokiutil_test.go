package lokiutil

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestWrapEntryLog(t *testing.T) {
	entries := []streamEntry{
		{Log: "W0627 19:40:45.997742\t 1 warnings.go:70] <msg> duplicate name \"PPS_PIPELINE_NAME\"\n"},
		{Log: "{\"message\":\"started setting up External Pachd GRPC Server\",\"severity\":\"info\"}"},
	}
	for _, entry := range entries {
		t.Run("test wrapEntryLog", func(t *testing.T) {
			wrappedEntry, err := wrapEntryLog(entry)
			require.NoError(t, err, "entries should be serializable")
			require.True(t, strings.HasPrefix(wrappedEntry.Log, "{"), "wrappedEntry.Log should be JSON")
		})
	}
}
