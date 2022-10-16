//go:build k8s

package cmds

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// loginAsUser sets the auth token in the pachctl config to a token for `user`
func loginAsUser(t *testing.T, c *client.APIClient, user string) {
	configPath := executeCmdAndGetLastWord(t, tu.PachctlBashCmd(t, c, `echo $PACH_CONFIG`))
	if user == auth.RootUser {
		require.NoError(t, config.WritePachTokenToConfigPath(tu.RootToken, configPath, false))
		return
	}
	rootClient := tu.AuthenticatedPachClient(t, c, auth.RootUser)
	robot := strings.TrimPrefix(user, auth.RobotPrefix)
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: robot})
	require.NoError(t, err)
	require.NoError(t, config.WritePachTokenToConfigPath(token.Token, configPath, false))
}

// this function executes the command and returns the last word
// in the output
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

// TestActivate tests that activating, deactivating and re-activating works.
// This means all cluster state is being reset correctly.
func TestActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateEnterprise(t, c)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
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
	tu.ActivateEnterprise(t, c)
	clientId := tu.UniqueString("clientId")
	// activation fails to activate with bad issuer URL
	require.YesError(t, tu.PachctlBashCmd(t, c, `
		echo '{{.token}}' | pachctl auth activate --issuer 'bad-url.com' --client-id {{.id}} --supply-root-token`,
		"token", tu.RootToken,
		"id", clientId,
	).Run())

	// the OIDC client does not exist in pachd
	require.YesError(t, tu.PachctlBashCmd(t, c, `
		pachctl idp list-client | match '{{.id}}'`,
		"id", clientId,
	).Run())

	// activation succeeds when passed happy-path values
	require.NoError(t, tu.PachctlBashCmd(t, c, `
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

	cmd := tu.PachctlBashCmd(t, c, "pachctl auth login --no-browser")
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)

	c = tu.UnauthenticatedPachClient(t, c)
	require.NoError(t, cmd.Start())
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		if strings.HasPrefix(strings.TrimSpace(sc.Text()), "http://") {
			tu.DoOAuthExchange(t, c, c, sc.Text())
			break
		}
	}
	require.NoError(t, cmd.Wait())
	require.NoError(t, tu.PachctlBashCmd(t, c, `
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
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo '{{.token}}' | pachctl auth login --id-token
		pachctl auth whoami | match user:{{.user}}`,
		"user", tu.DexMockConnectorEmail,
		"token", token,
	).Run())
}

func TestWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	alice := tu.UniqueString("robot:alice")
	loginAsUser(t, c, alice)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami | match {{.alice}}`,
		"alice", alice,
	).Run())
}

func TestCheckGetSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	alice, bob := tu.UniqueString("robot:alice"), tu.UniqueString("robot:bob")
	// Test both forms of the 'pachctl auth get' command, as well as 'pachctl auth check'

	loginAsUser(t, c, alice)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create repo {{.repo}}
		pachctl auth check repo {{.repo}} \
                        | match 'Roles: \[repoOwner\]'
		pachctl auth get repo {{.repo}} \
			| match {{.alice}}
		`,
		"alice", alice,
		"bob", bob,
		"repo", tu.UniqueString("TestGet-repo"),
	).Run())

	repo := tu.UniqueString("TestGet-repo")
	// Test 'pachctl auth set'
	require.NoError(t, tu.PachctlBashCmd(t, c, `pachctl create repo {{.repo}}
		pachctl auth set repo {{.repo}} repoReader {{.bob}}
		pachctl auth get repo {{.repo}}\
			| match "{{.bob}}: \[repoReader\]" \
			| match "{{.alice}}: \[repoOwner\]"
		`,
		"alice", alice,
		"bob", bob,
		"repo", repo,
	).Run())

	// Test checking another user's permissions
	loginAsUser(t, c, auth.RootUser)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth check repo {{.repo}} {{.alice}} \
			| match "Roles: \[repoOwner\]"
                pachctl auth check repo {{.repo}} {{.bob}} \
			| match "Roles: \[repoReader\]"
		`,
		"alice", alice,
		"bob", bob,
		"repo", repo,
	).Run())
}

func TestAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)
	// Modify the list of admins to add 'admin2'
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth get cluster \
			| match "pach:root"
		pachctl auth set cluster clusterAdmin robot:admin
		pachctl auth set cluster clusterAdmin robot:admin2
		pachctl auth get cluster \
			| match "^robot:admin2: \[clusterAdmin\]$" \
			| match "^robot:admin: \[clusterAdmin\]$"
		pachctl auth set cluster none robot:admin

		# as 'admin' is a substr of 'admin2', use '^admin$' regex...
		pachctl auth get cluster \
			| match -v "^robot:admin$" \
			| match "^robot:admin2: \[clusterAdmin\]$"
		`).Run())

	// Now 'admin2' is the only admin. Login as admin2, and swap 'admin' back in
	// (so that deactivateAuth() runs), and call 'list-admin' (to make sure it
	// works for non-admins)
	loginAsUser(t, c, "robot:admin2")
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth set cluster clusterAdmin robot:admin
		pachctl auth set cluster none robot:admin2
	`).Run())
	require.NoError(t, backoff.Retry(func() error {
		return errors.EnsureStack(tu.PachctlBashCmd(t, c, `pachctl auth get cluster \
			| match -v "robot:admin2" \
			| match "robot:admin"
		`).Run())
	}, backoff.NewTestingBackOff()))
}

func TestGetAndUseRobotToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)
	// Test both get-robot-token and use-auth-token; make sure that they work
	// together with -q
	require.NoError(t, tu.PachctlBashCmd(t, c, `pachctl auth get-robot-token -q marvin \
	  | pachctl auth use-auth-token
	pachctl auth whoami \
	  | match 'robot:marvin'
		`).Run())
}

func TestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, false))

	require.NoError(t, tu.PachctlBashCmd(t, c, `
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

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth get-config -o yaml \
		  | match 'issuer: http://pachd:1658/dex' \
		  | match 'localhost_issuer: true' \
		  | match 'client_id: localhost' \
		  | match 'redirect_uri: http://localhost:1650' \
		`).Run())
}

// TestGetRobotTokenTTL tests that the --ttl argument to 'pachctl get-robot-token'
// correctly limits the lifetime of the returned token
func TestGetRobotTokenTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)

	alice := tu.UniqueString("alice")

	var tokenBuf bytes.Buffer
	tokenCmd := tu.PachctlBashCmd(t, c, `pachctl auth get-robot-token {{.alice}} --ttl=1h -q`, "alice", alice)
	tokenCmd.Stdout = &tokenBuf
	require.NoError(t, tokenCmd.Run())
	token := strings.TrimSpace(tokenBuf.String())

	login := tu.PachctlBashCmd(t, c, `echo {{.token}} | pachctl auth use-auth-token
		pachctl auth whoami | \
		match 'session expires: '
	`, "token", token)
	require.NoError(t, login.Run())
}

// TestGetOwnGroups tests that calling `pachctl auth get-groups` with no arguments
// returns the current user's groups.
func TestGetOwnGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	rootClient := tu.AuthenticatedPachClient(t, c, auth.RootUser)

	group := tu.UniqueString("group")

	_, err := rootClient.ModifyMembers(rootClient.Ctx(), &auth.ModifyMembersRequest{
		Group: group,
		Add:   []string{auth.RootUser}},
	)
	require.NoError(t, err)

	require.NoError(t, tu.PachctlBashCmd(t, rootClient, `pachctl auth get-groups | match '{{ .group }}'`,
		"group", group).Run())
}

// TestGetGroupsForUser tests that calling `pachctl auth get-groups` with an argument
// returns the groups for the specified user.
func TestGetGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	rootClient := tu.AuthenticatedPachClient(t, c, auth.RootUser)

	alice := auth.RobotPrefix + tu.UniqueString("alice")
	group := tu.UniqueString("group")

	_, err := rootClient.ModifyMembers(rootClient.Ctx(), &auth.ModifyMembersRequest{
		Group: group,
		Add:   []string{alice}},
	)
	require.NoError(t, err)

	require.NoError(t, tu.PachctlBashCmd(t, c, `pachctl auth get-groups {{ .alice }} | match '{{ .group }}'`,
		"group", group, "alice", alice).Run())
}

// TestRotateRootToken tests that calling 'pachctl auth rotate-root-token' rotates the root user's token
func TestRotateRootToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	c = tu.AuthenticatedPachClient(t, c, auth.RootUser)

	sessionToken := c.AuthToken()

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami | match "pach:root"
	`).Run())

	// rotate current user's token
	token := executeCmdAndGetLastWord(t, tu.PachctlBashCmd(t, c, "pachctl auth rotate-root-token"))

	// current user (root) can't authenticate
	require.YesError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami
	`).Run())

	// root can authenticate once the new token is set
	configPath := executeCmdAndGetLastWord(t, tu.PachctlBashCmd(t, c, `echo $PACH_CONFIG`))
	require.NoError(t, config.WritePachTokenToConfigPath(token, configPath, false))
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami | match "pach:root"
	`).Run())

	// rotate to new token and get (the same) output token
	tu.PachctlBashCmd(t, c, "pachctl auth rotate-root-token --supply-token {{ .tok }}", "tok", sessionToken)

	token = executeCmdAndGetLastWord(t, tu.PachctlBashCmd(t, c, `
		pachctl auth rotate-root-token --supply-token {{ .tok }}`,
		"tok", sessionToken))

	require.Equal(t, sessionToken, token)

	require.YesError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami
	`).Run())

	// root can authenticate once the new token is set
	require.NoError(t, config.WritePachTokenToConfigPath(sessionToken, configPath, false))
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami | match "pach:root"
	`).Run())
}

// TestSynonyms walks through the command tree for each resource and verb combination defined in PPS.
// A template is filled in that calls the help flag and the output is compared. It seems like 'match'
// is unable to compare the outputs correctly, but we can use diff here which returns an exit code of 0
// if there is no difference.
func TestSynonyms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	synonymCheckTemplate := `
		pachctl auth {{VERB}} {{RESOURCE_SYNONYM}} -h > synonym.txt
		pachctl auth {{VERB}} {{RESOURCE}} -h > singular.txt
		diff synonym.txt singular.txt
		rm synonym.txt singular.txt
	`

	resources := map[string][]string{
		"repo": {"check", "set", "get"},
	}

	synonymsMap := map[string]string{
		"repo": "repos",
	}

	for resource, verbs := range resources {
		withResource := strings.ReplaceAll(synonymCheckTemplate, "{{RESOURCE}}", resource)
		withResources := strings.ReplaceAll(withResource, "{{RESOURCE_SYNONYM}}", synonymsMap[resource])

		for _, verb := range verbs {
			synonymCommand := strings.ReplaceAll(withResources, "{{VERB}}", verb)
			t.Logf("Testing auth %s %s -h\n", verb, resource)
			require.NoError(t, tu.BashCmd(synonymCommand).Run())
		}
	}
}

// TestRevokeToken tests revoking an existing token
func TestRevokeToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	root := tu.AuthenticatedPachClient(t, c, auth.RootUser)
	aliceName := auth.RobotPrefix + tu.UniqueString("alice")
	alice := tu.AuthenticateClient(t, c, aliceName)

	whoAmIResp, err := alice.WhoAmI(alice.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, aliceName, whoAmIResp.Username)

	require.NoError(t, tu.PachctlBashCmd(t, root,
		`pachctl auth revoke --token={{.alice_token}}`,
		"alice_token", alice.AuthToken()).Run())

	_, err = alice.WhoAmI(alice.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, err)
	require.True(t, auth.IsErrBadToken(err))
}

// TestRevokeUser tests revoking all tokens currently issues for a user
func TestRevokeUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	root := tu.AuthenticatedPachClient(t, c, auth.RootUser)
	aliceName := auth.RobotPrefix + tu.UniqueString("alice")
	aliceClients := make([]*client.APIClient, 3)
	for i := 0; i < len(aliceClients); i++ {
		aliceClients[i] = tu.AuthenticateClient(t, c, aliceName)
	}

	for i := 0; i < len(aliceClients); i++ {
		c := aliceClients[i]
		whoAmIResp, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.NoError(t, err)
		require.Equal(t, aliceName, whoAmIResp.Username)
	}

	require.NoError(t, tu.PachctlBashCmd(t, root,
		`pachctl auth revoke --user={{.alice}}`,
		"alice", aliceName).Run())

	// See relevant comments in TestRevokeToken
	for i := 0; i < len(aliceClients); i++ {
		c := aliceClients[i]
		_, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.YesError(t, err)
		require.True(t, auth.IsErrBadToken(err))
	}
}
