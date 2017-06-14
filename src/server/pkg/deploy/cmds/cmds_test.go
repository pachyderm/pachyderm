package cmds

import (
	"os/exec"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestDashImageExists(t *testing.T) {
	c := exec.Command("docker", "pull", defaultDashImage)
	require.NoError(t, c.Run())
}
