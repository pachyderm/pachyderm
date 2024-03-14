//go:build k8s

package cmds

import (
	"bufio"
	"os/exec"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestGetLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	// TODO(CORE-2123): check for real logs
	require.NoError(t, testutil.PachctlBashCmd(t, c, `
		pachctl logs2 | match "GetLogs dummy response"`,
	).Run())
}

// loginAsUser sets the auth token in the pachctl config to a token for `user`.
// Stolen from auth/cmds tests.
func loginAsUser(t *testing.T, c *client.APIClient, user string) {
	t.Helper()
	configPath := executeCmdAndGetLastWord(t, testutil.PachctlBashCmd(t, c, `echo $PACH_CONFIG`))
	rootClient := testutil.AuthenticatedPachClient(t, c, auth.RootUser)
	if user == auth.RootUser {
		require.NoError(t, config.WritePachTokenToConfigPath(testutil.RootToken, configPath, false))
		return
	}
	robot := strings.TrimPrefix(user, auth.RobotPrefix)
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: robot})
	require.NoError(t, err)
	require.NoError(t, config.WritePachTokenToConfigPath(token.Token, configPath, false))
}

// executeCmdAndGetLastWord executes an exec.Cmd and returns the last word in
// the output.  Stolen from auth/cmds tests.
func executeCmdAndGetLastWord(t *testing.T, cmd *exec.Cmd) string {
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	sc := bufio.NewScanner(out)
	sc.Split(bufio.ScanWords)
	var token string
	for sc.Scan() {
		tmp := sc.Text()
		if strings.TrimSpace(tmp) != "" {
			token = tmp
		}
	}
	require.NoError(t, cmd.Wait())
	return token
}

func TestGetLogs_nonadmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	alice := testutil.UniqueString("robot:alice")
	loginAsUser(t, c, alice)
	// TODO(CORE-2123): check for real logs
	require.NoError(t, testutil.PachctlBashCmd(t, c, `
		pachctl logs2 | match "GetLogs dummy response"`,
	).Run())
}
