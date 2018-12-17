package fruit_stand

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestFruitStand(t *testing.T) {
	require.NoError(t, exec.Command("pachctl", "create-repo", "data").Run())
	require.NoError(t, exec.Command("pachctl", "start-commit", "data", "master").Run())
	putFileCmd := exec.Command("pachctl", "put-file", "data", "master", "sales")
	set1, err := os.Open("set1.txt")
	require.NoError(t, err)
	putFileCmd.Stdin = set1
	require.NoError(t, putFileCmd.Run())
	require.NoError(t, exec.Command("pachctl", "finish-commit", "data", "master").Run())
	require.NoError(t, exec.Command("pachctl", "create-pipeline", "-f", "pipeline.json").Run())
	time.Sleep(5 * time.Second)
	require.NoError(t, exec.Command("pachctl", "flush-commit", "data/master").Run())
}
