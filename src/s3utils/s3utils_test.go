package s3utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestForEachFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip()
	}
	var files []string
	err := ForEachFile("s3://pachyderm-test/pipeline", true, "", func(file string, modtime time.Time) error {
		files = append(files, file)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{"pipeline/file"}, files)
}
