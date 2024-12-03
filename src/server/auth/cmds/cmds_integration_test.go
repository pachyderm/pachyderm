//go:build k8s

package cmds

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// TestActivate tests that activating, deactivating and re-activating works.
// This means all cluster state is being reset correctly.
func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		echo '{{.token}}' | pachctl auth activate --supply-root-token
		pachctl auth whoami | match {{.user}}
		echo 'y' | pachctl auth deactivate
		echo '{{.token}}' | pachctl auth activate --supply-root-token
		pachctl auth whoami | match {{.user}}
		echo 'y' | pachctl auth deactivate`,
		"token", tu.RootToken,
		"user", auth.RootUser,
	).Run())
}

// TestActivateFailureRollback tests that any partial state left
// from a failed execution is cleaned up
func TestActivateFailureRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	clientId := uuid.UniqueString("clientId")
	// activation fails to activate with bad issuer URL
	require.YesError(t, testutilpachctl.PachctlBashCmd(t, c, `
		echo '{{.token}}' | pachctl auth activate --issuer 'bad-url.com' --client-id {{.id}} --supply-root-token`,
		"token", tu.RootToken,
		"id", clientId,
	).Run())

	// the OIDC client does not exist in pachd
	require.YesError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl idp list-client | match '{{.id}}'`,
		"id", clientId,
	).Run())

	// activation succeeds when passed happy-path values
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		echo '{{.token}}' | pachctl auth activate --client-id {{.id}} --supply-root-token
		pachctl auth whoami | match {{.user}}`,
		"token", tu.RootToken,
		"user", auth.RootUser,
		"id", clientId,
	).Run())
}

func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)

	// Configure OIDC login
	require.NoError(t, tu.ConfigureOIDCProvider(t, tu.AuthenticateClient(t, c, auth.RootUser), false))

	require.NoErrorWithinTRetryConstant(t, 5*time.Minute, func() error {
		ctx, done := context.WithTimeout(pctx.Background("auth.login"), 30*time.Second)
		defer done()
		cmd := testutilpachctl.PachctlBashCmdCtx(ctx, t, c, "echo '' | pachctl auth use-auth-token && pachctl auth login --no-browser")
		out, err := cmd.StdoutPipe()
		if err != nil {
			return errors.Wrap(err, "StdoutPipe")
		}
		var buf bytes.Buffer
		cmd.Stderr = &buf

		c := tu.UnauthenticatedPachClient(t, c)
		if err := cmd.Start(); err != nil {
			return errors.Wrap(err, "cmd.Start")
		}
		sc := bufio.NewScanner(out)
		for sc.Scan() {
			if strings.HasPrefix(strings.TrimSpace(sc.Text()), "http://") {
				url := sc.Text()
				t.Logf("doing OAuth exchange against %v", url)
				if err := tu.DoOAuthExchangeOnce(t, c, c, url); err != nil {
					return errors.Wrap(err, "DoOAuthExchangeOnce")
				}
				break
			}
		}
		if err := cmd.Wait(); err != nil {
			return errors.Wrap(err, "cmd.Wait")
		}
		require.False(t, strings.Contains(buf.String(), "Could not inspect"), "does not inspect project when auth is enabled and no credentials are provided")
		return nil
	}, time.Second, "should pachctl auth login")
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl auth whoami | match user:{{.user}}`,
		"user", tu.DexMockConnectorEmail,
	).Run())
}

func TestLoginIDToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	c = tu.AuthenticateClient(t, c, auth.RootUser)
	// Configure OIDC login
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, false))

	// Get an ID token for a trusted peer app
	token := tu.GetOIDCTokenForTrustedApp(t, c, false)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		echo '{{.token}}' | pachctl auth login --id-token
		pachctl auth whoami | match user:{{.user}}`,
		"user", tu.DexMockConnectorEmail,
		"token", token,
	).Run())
}

func TestIDTokenFromEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	c = tu.AuthenticateClient(t, c, auth.RootUser)
	// Configure OIDC login
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, false))
	// Get an ID token for a trusted peer app
	token := tu.GetOIDCTokenForTrustedApp(t, c, false)
	require.YesError(t, testutilpachctl.PachctlBashCmd(t, c, `
                echo "" | pachctl auth use-auth-token;
		pachctl auth whoami`,
	).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
                echo "" | pachctl auth use-auth-token;
                export DEX_TOKEN={{.token}};
                pachctl auth whoami | match user:{{.user}}`,
		"user", tu.DexMockConnectorEmail,
		"token", token,
	).Run())
}

func TestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, false))

	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
        pachctl auth set-config <<EOF
        {
            "issuer": "http://pachd:1658/dex",
            "localhost_issuer": true,
            "client_id": "localhost",
            "redirect_uri": "http://localhost:1650"
        }
EOF
		pachctl auth get-config \
		  | match '"issuer": "http://pachd:1658/dex"' \
		  | match '"localhost_issuer": true' \
		  | match '"client_id": "localhost"' \
		  | match '"redirect_uri": "http://localhost:1650"' \
		  | match '}'
		`).Run())

	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl auth get-config -o yaml \
		  | match 'issuer: http://pachd:1658/dex' \
		  | match 'localhost_issuer: true' \
		  | match 'client_id: localhost' \
		  | match 'redirect_uri: http://localhost:1650' \
		`).Run())
}

// TestRotateRootToken tests that calling 'pachctl auth rotate-root-token' rotates the root user's token
func TestRotateRootToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)

	sessionToken := c.AuthToken()

	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
               pachctl auth whoami | match "pach:root"
       `).Run())

	// rotate current user's token
	token := executeCmdAndGetLastWord(t, testutilpachctl.PachctlBashCmd(t, c, "pachctl auth rotate-root-token"))

	// current user (root) can't authenticate
	require.YesError(t, testutilpachctl.PachctlBashCmd(t, c, `
               pachctl auth whoami
       `).Run())

	// root can authenticate once the new token is set
	configPath := executeCmdAndGetLastWord(t, testutilpachctl.PachctlBashCmd(t, c, `echo $PACH_CONFIG`))
	require.NoError(t, config.WritePachTokenToConfigPath(token, configPath, false))
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
               pachctl auth whoami | match "pach:root"
       `).Run())

	// rotate to new token and get (the same) output token
	testutilpachctl.PachctlBashCmd(t, c, "pachctl auth rotate-root-token --supply-token {{ .tok }}", "tok", sessionToken)

	token = executeCmdAndGetLastWord(t, testutilpachctl.PachctlBashCmd(t, c, `
               pachctl auth rotate-root-token --supply-token {{ .tok }}`,
		"tok", sessionToken))

	require.Equal(t, sessionToken, token)

	require.YesError(t, testutilpachctl.PachctlBashCmd(t, c, `
               pachctl auth whoami
       `).Run())

	// root can authenticate once the new token is set
	require.NoError(t, config.WritePachTokenToConfigPath(sessionToken, configPath, false))
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
               pachctl auth whoami | match "pach:root"
       `).Run())
}
