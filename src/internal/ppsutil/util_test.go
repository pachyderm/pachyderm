package ppsutil_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
)

func TestPipelineRcName(t *testing.T) {
	for _, test := range []struct {
		name    string
		version uint64
		result  string
	}{
		{"foo", 1, "pipeline-foo-v1"},
		{"012345678901234567890123456789012345678901234567890123456789", 1, "pipeline-012345678901234567890123456789012345678901234567890-v1"},
		{"012345678901234567890123456789012345678901234567890123456789", 1024, "pipeline-012345678901234567890123456789012345678901234567-v1024"},
	} {
		if result := ppsutil.PipelineRcName(test.name, test.version); result != test.result {
			t.Errorf("PipelineRcName(%q, %d) = %q, not %q", test.name, test.version, result, test.result)
		} else if len(result) > 63 {
			t.Errorf("len(PipelineRcName(%q, %d)) = %d", test.name, test.version, len(result))
		}
	}
}
