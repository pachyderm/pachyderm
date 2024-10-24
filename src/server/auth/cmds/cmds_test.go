package cmds

import (
	"bufio"
	"bytes"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"os/exec"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// executeCmdAndGetLastWord executes the command and returns the last word in the output.
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

// loginAsUser sets the auth token in the pachctl config to a token for `user`
func loginAsUser(t *testing.T, rootClient *client.APIClient, user string) {
	t.Helper()
	configPath := executeCmdAndGetLastWord(t, testutilpachctl.PachctlBashCmdCtx(rootClient.Ctx(), t, rootClient, `echo $PACH_CONFIG`))
	if user == auth.RootUser {
		require.NoError(t, config.WritePachTokenToConfigPath(tu.RootToken, configPath, false))
		return
	}
	robot := strings.TrimPrefix(user, auth.RobotPrefix)
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: robot})
	require.NoError(t, err)
	require.NoError(t, config.WritePachTokenToConfigPath(token.Token, configPath, false))
}

func TestWhoAmI(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))
	alice := uuid.UniqueString("robot:alice")
	loginAsUser(t, c, alice)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl auth whoami | match {{.alice}}`,
		"alice", alice,
	).Run())
}

// TestCheckGetSetRepo tests 3 pachctl auth subcommands: check, get, and set, on repos.
// Test both modes of `check repo`: 1) check caller's own permissions 2) check another user's permissions.
func TestCheckGetSetRepo(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))

	// alice sets up their own project and repos so that they can run the appropriate auth commands below.
	alice, bob := uuid.UniqueString("robot:alice"), uuid.UniqueString("robot:bob")
	loginAsUser(t, c, alice)
	project, repo := uuid.UniqueString("project"), uuid.UniqueString("repo")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl create project {{.project}}`, "project", project).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl create repo --project {{.project}} {{.repo}}`, "project", pfs.DefaultProjectName, "repo", repo).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl create repo --project {{.project}} {{.repo}}`, "project", project, "repo", repo).Run())

	// alice can check, get, and set permissions wrt their own repo
	// but they can't check other users' permissions because they lack CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL
	for _, project := range []string{pfs.DefaultProjectName, project} {
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `cat $PACH_CONFIG >/tmp/bazquux`,
			"project", project,
			"repo", repo,
		).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
pachctl auth check repo {{.repo}} --project {{.project}} | match repoOwner
pachctl auth check repo {{.repo}} --project {{.project}} >/tmp/bim 2>&1`,
			"project", project,
			"repo", repo,
		).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get repo {{.repo}} --project {{.project}} >/tmp/foobar 2>&1`,
			"project", project,
			"repo", repo,
			"alice", alice).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get repo {{.repo}} --project {{.project}} | match "{{.alice}}: \[repoOwner\]"`,
			"project", project,
			"repo", repo,
			"alice", alice,
		).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth set repo {{.repo}} repoReader {{.bob}} --project {{.project}}`,
			"project", project,
			"repo", repo,
			"bob", bob,
		).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get repo {{.repo}} --project {{.project}} | match "{{.bob}}: \[repoReader\]"`,
			"project", project,
			"repo", repo,
			"bob", bob,
		).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
			pachctl auth check repo {{.repo}} --project {{.project}} | match repoOwner
			pachctl auth get repo {{.repo}} --project {{.project}} | match "{{.alice}}: \[repoOwner\]"
			pachctl auth set repo {{.repo}} repoReader {{.bob}} --project {{.project}}
			pachctl auth get repo {{.repo}} --project {{.project}} | match "{{.bob}}: \[repoReader\]"
			`,
			"project", project,
			"repo", repo,
			"alice", alice,
			"bob", bob,
		).Run())

		require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth check repo {{.repo}} {{.user}} --project {{.project}}`, "project", project, "repo", repo, "user", alice).Run())
	}

	// root user can check everyone's role bindings because they have CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL
	loginAsUser(t, c, auth.RootUser)
	for _, project := range []string{pfs.DefaultProjectName, project} {
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth check repo {{.repo}} {{.user}} --project {{.project}} | match repoOwner`, "project", project, "repo", repo, "user", alice).Run())
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth check repo {{.repo}} {{.user}} --project {{.project}} | match repoReader`, "project", project, "repo", repo, "user", bob).Run())
	}
}

func TestCheckGetSetProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project project
		pachctl auth check project project | match clusterAdmin | match projectOwner
		pachctl auth get project project | match projectOwner
		pachctl auth set project project repoReader,projectOwner pach:root
		pachctl auth get project project | match projectOwner | match repoReader

		pachctl auth get robot-auth-token alice
		pachctl auth check project project robot:alice | match "Roles: \[\]"
		pachctl auth set project project projectOwner robot:alice
		pachctl auth check project project robot:alice | match "projectOwner"
	`).Run())
}

func TestAdmins(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))
	// Modify the list of admins to add 'admin2'
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
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
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl auth set cluster clusterAdmin robot:admin
		pachctl auth set cluster none robot:admin2
	`).Run())
	require.NoError(t, backoff.Retry(func() error {
		return errors.EnsureStack(testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get cluster \
			| match -v "robot:admin2" \
			| match "robot:admin"
		`).Run())
	}, backoff.NewTestingBackOff()))
}

func TestGetAndUseRobotToken(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))
	// Test both get-robot-token and use-auth-token; make sure that they work
	// together with -q
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get-robot-token -q marvin \
	  | pachctl auth use-auth-token
	pachctl auth whoami \
	  | match 'robot:marvin'
		`).Run())
}

// TestGetRobotTokenTTL tests that the --ttl argument to 'pachctl get-robot-token'
// correctly limits the lifetime of the returned token
func TestGetRobotTokenTTL(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))

	alice := uuid.UniqueString("alice")
	var tokenBuf bytes.Buffer
	tokenCmd := testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get-robot-token {{.alice}} --ttl=1h -q`, "alice", alice)
	tokenCmd.Stdout = &tokenBuf
	require.NoError(t, tokenCmd.Run())
	token := strings.TrimSpace(tokenBuf.String())

	login := testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `echo {{.token}} | pachctl auth use-auth-token
		pachctl auth whoami | \
		match 'session expires: '
	`, "token", token)
	require.NoError(t, login.Run())
}

// TestGetOwnGroups tests that calling `pachctl auth get-groups` with no arguments
// returns the current user's groups.
func TestGetOwnGroups(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))

	group := uuid.UniqueString("group")

	_, err := c.ModifyMembers(c.Ctx(), &auth.ModifyMembersRequest{
		Group: group,
		Add:   []string{auth.RootUser}},
	)
	require.NoError(t, err)

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get-groups | match '{{ .group }}'`,
		"group", group).Run())
}

// TestGetGroupsForUser tests that calling `pachctl auth get-groups` with an argument
// returns the groups for the specified user.
func TestGetGroups(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))

	alice := auth.RobotPrefix + uuid.UniqueString("alice")
	group := uuid.UniqueString("group")

	_, err := c.ModifyMembers(c.Ctx(), &auth.ModifyMembersRequest{
		Group: group,
		Add:   []string{alice}},
	)
	require.NoError(t, err)

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl auth get-groups {{ .alice }} | match '{{ .group }}'`,
		"group", group, "alice", alice).Run())
}

// TestSynonyms walks through the command tree for each resource and verb combination defined in PPS.
// A template is filled in that calls the help flag and the output is compared. It seems like 'match'
// is unable to compare the outputs correctly, but we can use diff here which returns an exit code of 0
// if there is no difference.
func TestSynonyms(t *testing.T) {
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
			require.NoError(t, testutilpachctl.BashCmd(synonymCommand).Run())
		}
	}
}

// TestRevokeToken tests revoking an existing token
func TestRevokeToken(t *testing.T) {
	ctx := pctx.TestContext(t)
	root := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))

	aliceName := auth.RobotPrefix + uuid.UniqueString("alice")
	alice := tu.AuthenticateClient(t, root, aliceName)

	whoAmIResp, err := alice.WhoAmI(alice.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, aliceName, whoAmIResp.Username)

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, root, `
		pachctl auth revoke --token={{.alice_token}} | match '1 auth token revoked'
		pachctl auth revoke --token={{.alice_token}} | match '0 auth tokens revoked'
		`,
		"alice_token", alice.AuthToken()).Run())

	_, err = alice.WhoAmI(alice.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, err)
	require.True(t, auth.IsErrBadToken(err))
}

// TestRevokeUser tests revoking all tokens currently issues for a user
func TestRevokeUser(t *testing.T) {
	ctx := pctx.TestContext(t)
	root := pachd.NewTestPachd(t, pachd.ActivateAuthOption(tu.RootToken))

	aliceName := auth.RobotPrefix + uuid.UniqueString("alice")
	aliceClients := make([]*client.APIClient, 3)
	for i := 0; i < len(aliceClients); i++ {
		aliceClients[i] = tu.AuthenticateClient(t, root, aliceName)
	}

	for i := 0; i < len(aliceClients); i++ {
		c := aliceClients[i]
		whoAmIResp, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.NoError(t, err)
		require.Equal(t, aliceName, whoAmIResp.Username)
	}

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, root, `
		pachctl auth revoke --user={{.alice}} | match '3 auth tokens revoked'
		pachctl auth revoke --user={{.alice}} | match '0 auth tokens revoked'
		`,
		"alice", aliceName).Run())

	// See relevant comments in TestRevokeToken
	for i := 0; i < len(aliceClients); i++ {
		c := aliceClients[i]
		_, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.YesError(t, err)
		require.True(t, auth.IsErrBadToken(err))
	}
}
