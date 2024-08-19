package scraper

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScraper(t *testing.T) {
	t.Skip("1.4 example not updated for minikubetestenv")
	require.NoError(t, exec.Command("pachctl", "create", "repo", "urls").Run())
	require.NoError(t, exec.Command("pachctl", "start", "commit", "urls@master").Run())
	putFileCmd := exec.Command("pachctl", "put", "file", "urls@master:urls")
	urls, err := os.Open("urls")
	require.NoError(t, err)
	putFileCmd.Stdin = urls
	require.NoError(t, putFileCmd.Run())
	require.NoError(t, exec.Command("pachctl", "finish", "commit", "urls@master").Run())
	require.NoError(t, exec.Command("pachctl", "create", "pipeline", "-f", "scraper.json").Run())
}
