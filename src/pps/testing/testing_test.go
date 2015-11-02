package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
)

func TestSimple(t *testing.T) {
	t.Skip()
	RunTest(t, testSimple)
}

func testSimple(t *testing.T, pfsAPIClient pfs.APIClient, jobAPIClient pps.JobAPIClient, pipelineAPIClient pps.PipelineAPIClient) {
}
