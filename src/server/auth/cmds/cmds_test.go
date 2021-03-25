package cmds

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// loginAsUser sets the auth token in the pachctl config to a token for `user`
func loginAsUser(t *testing.T, user string) {
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)
	robot := strings.TrimPrefix(user, auth.RobotPrefix)
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: robot})
	require.NoError(t, err)
	config.WritePachTokenToConfig(token.Token)
}

func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	// Configure OIDC login
	tu.ConfigureOIDCProvider(t)

	cmd := exec.Command("pachctl", "auth", "login", "--no-browser")
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		if strings.HasPrefix(strings.TrimSpace(sc.Text()), "http://") {
			tu.DoOAuthExchange(t, sc.Text())
			break
		}
	}
	cmd.Wait()

	require.NoError(t, tu.BashCmd(`
		pachctl auth whoami | match user:{{.user}}`,
		"user", tu.DexMockConnectorEmail,
	).Run())
}

func TestLoginIDToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	// Configure OIDC login
	tu.ConfigureOIDCProvider(t)

	// Get an ID token for a trusted peer app
	token := tu.GetOIDCTokenForTrustedApp(t)

	require.NoError(t, tu.BashCmd(`
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
	alice := tu.UniqueString("robot:alice")
	loginAsUser(t, alice)
	defer tu.DeleteAll(t)
	require.NoError(t, tu.BashCmd(`
		pachctl auth whoami | match {{.alice}}`,
		"alice", alice,
	).Run())
}

func TestCheckGetSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	defer tu.DeleteAll(t)

	alice, bob := tu.UniqueString("robot:alice"), tu.UniqueString("robot:bob")
	// Test both forms of the 'pachctl auth get' command, as well as 'pachctl auth check'

	loginAsUser(t, alice)
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}}
		pachctl auth check repo REPO_MODIFY_BINDINGS {{.repo}}
		pachctl auth get repo {{.repo}} \
			| match {{.alice}}
		`,
		"alice", alice,
		"bob", bob,
		"repo", tu.UniqueString("TestGet-repo"),
	).Run())

	// Test 'pachctl auth set'
	require.NoError(t, tu.BashCmd(`pachctl create repo {{.repo}}
		pachctl auth set repo {{.repo}} repoReader {{.bob}}
		pachctl auth get repo {{.repo}}\
			| match "{{.bob}}: \[repoReader\]" \
			| match "{{.alice}}: \[repoOwner\]"
		`,
		"alice", alice,
		"bob", bob,
		"repo", tu.UniqueString("TestGet-repo"),
	).Run())
}

func TestAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	// Modify the list of admins to add 'admin2'
	require.NoError(t, tu.BashCmd(`
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
	loginAsUser(t, "robot:admin2")
	require.NoError(t, tu.BashCmd(`
		pachctl auth set cluster clusterAdmin robot:admin 
		pachctl auth set cluster none robot:admin2
	`).Run())
	require.NoError(t, backoff.Retry(func() error {
		return tu.BashCmd(`pachctl auth get cluster \
			| match -v "robot:admin2" \
			| match "robot:admin"
		`).Run()
	}, backoff.NewTestingBackOff()))
}

func TestGetAndUseAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	// Test both get-auth-token and use-auth-token; make sure that they work
	// together with -q
	require.NoError(t, tu.BashCmd(`pachctl auth get-auth-token -q robot:marvin \
	  | pachctl auth use-auth-token
	pachctl auth whoami \
	  | match 'robot:marvin'
		`).Run())
}

func TestGetAndUseRobotToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	// Test both get-robot-token and use-auth-token; make sure that they work
	// together with -q
	require.NoError(t, tu.BashCmd(`pachctl auth get-robot-token -q marvin \
	  | pachctl auth use-auth-token
	pachctl auth whoami \
	  | match 'robot:marvin'
		`).Run())
}

func TestConfig(t *testing.T) {
	if os.Getenv("RUN_BAD_TESTS") == "" {
		t.Skip("Skipping because RUN_BAD_TESTS was empty")
	}
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	require.NoError(t, tu.BashCmd(`
		pachctl auth set-config <<EOF
		{
		   "issuer": "http://localhost:658",
                   "localhost_issuer": true,
                   "client_id": localhost,
                   "redirect_uri": "http://localhost:650"
		}
		EOF
		pachctl auth get-config \
		  | match '"issuer": "localhost:658"' \
		  | match '"localhost_issuer": true,' \
		  | match '"client_id": "localhost"' \
		  | match '"redirect_uri": "http://localhost:650"' \
		  | match '}'
		`).Run())

	require.NoError(t, tu.BashCmd(`
		pachctl auth get-config -o yaml \
                  | match 'issuer: "localhost:658"' \
		  | match 'localhost_issuer: true,' \
		  | match 'client_id: localhost' \
		  | match 'redirect_uri: "http://localhost:650"' \
		`).Run())
}

// TestGetRobotTokenTTL tests that the --ttl argument to 'pachctl get-robot-token'
// correctly limits the lifetime of the returned token
func TestGetRobotTokenTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	alice := tu.UniqueString("alice")

	var tokenBuf bytes.Buffer
	tokenCmd := tu.BashCmd(`pachctl auth get-robot-token {{.alice}} --ttl=1h -q`, "alice", alice)
	tokenCmd.Stdout = &tokenBuf
	require.NoError(t, tokenCmd.Run())
	token := strings.TrimSpace(tokenBuf.String())

	login := tu.BashCmd(`echo {{.token}} | pachctl auth use-auth-token
		pachctl auth whoami | \
		match 'session expires: '
	`, "token", token)
	require.NoError(t, login.Run())
}

// TestGetAuthTokenTTL tests that the --ttl argument to 'pachctl get-auth-token'
// correctly limits the lifetime of the returned token
func TestGetAuthTokenTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	alice := tu.UniqueString("robot:alice")

	var tokenBuf bytes.Buffer
	tokenCmd := tu.BashCmd(`pachctl auth get-auth-token {{.alice}} --ttl=5s -q`, "alice", alice)
	tokenCmd.Stdout = &tokenBuf
	require.NoError(t, tokenCmd.Run())
	token := strings.TrimSpace(tokenBuf.String())

	time.Sleep(6 * time.Second)
	var errMsg bytes.Buffer
	login := tu.BashCmd(`echo {{.token}} | pachctl auth use-auth-token
		pachctl auth whoami
	`, "token", token)
	login.Stderr = &errMsg
	require.YesError(t, login.Run())
	require.Matches(t, "try logging in", errMsg.String())
}

func TestMain(m *testing.M) {
	// Preemptively deactivate Pachyderm auth (to avoid errors in early tests)
	if err := tu.BashCmd("echo 'iamroot' | pachctl auth use-auth-token &>/dev/null").Run(); err != nil {
		panic(err.Error())
	}

	if err := tu.BashCmd("pachctl auth whoami &>/dev/null").Run(); err == nil {
		if err := tu.BashCmd("yes | pachctl auth deactivate").Run(); err != nil {
			panic(err.Error())
		}
	}
	time.Sleep(time.Second)
	backoff.Retry(func() error {
		cmd := tu.Cmd("pachctl", "auth", "login")
		cmd.Stdin = strings.NewReader("admin\n")
		cmd.Stdout, cmd.Stderr = ioutil.Discard, ioutil.Discard
		if cmd.Run() != nil {
			return nil // cmd errored -- auth is deactivated
		}
		return errors.New("auth not deactivated yet")
	}, backoff.RetryEvery(time.Second))
	os.Exit(m.Run())
}
