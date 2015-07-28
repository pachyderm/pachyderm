package s3utils

import (
	"testing"
	"time"

	"go.pedge.io/protolog/logrus"

	"github.com/stretchr/testify/require"
)

func init() {
	logrus.Register()
}

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
