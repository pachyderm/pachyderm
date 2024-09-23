package worker_test

import (
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/internal/worker"
)

func TestProgramFS(t *testing.T) {
	var wfs = worker.ProgramFS()
	f, err := wfs.Open("name")
	require.NoError(t, err, "opening \"name\"")
	b, err := io.ReadAll(f)
	require.NoError(t, err, "reading \"name\"")
	require.Equal(t, "pps-worker", string(b))
}

func TestProgramHash(t *testing.T) {
	var wh = worker.ProgramHash()
	require.NotEqual(t, 0, len(wh), "program hash must have a value")
}
