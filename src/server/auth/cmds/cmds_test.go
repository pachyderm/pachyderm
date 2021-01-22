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

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// loginAsUser sets the auth token in the pachctl config to a token for `user`
func loginAsUser(t *testing.T, user string) {
	rootClient := tu.GetAuthenticatedPachClient(t, auth.RootUser)
	token, err := rootClient.GetAuthToken(rootClient.Ctx(), &auth.GetAuthTokenRequest{Subject: user})
	require.NoError(t, err)
	config.WritePachTokenToConfig(token.Token)
}

func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Configure OIDC login
	tu.ConfigureOIDCProvider(t)
	defer tu.DeleteAll(t)

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
		pachctl auth check owner {{.repo}}
		pachctl auth get {{.repo}} \
			| match {{.alice}}
		pachctl auth get {{.bob}} {{.repo}} \
			| match NONE
		`,
		"alice", alice,
		"bob", bob,
		"repo", tu.UniqueString("TestGet-repo"),
	).Run())

	// Test 'pachctl auth set'
	require.NoError(t, tu.BashCmd(`pachctl create repo {{.repo}}
		pachctl auth set {{.bob}} reader {{.repo}}
		pachctl auth get {{.bob}} {{.repo}} \
			| match READER
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
		pachctl auth list-admins \
			| match "pach:root"
		pachctl auth modify-admins --add robot:admin,robot:admin2
		pachctl auth list-admins \
			| match "^robot:admin2$" \
			| match "^robot:admin$" 
		pachctl auth modify-admins --remove robot:admin

		# as 'admin' is a substr of 'admin2', use '^admin$' regex...
		pachctl auth list-admins \
			| match -v "^robot:admin$" \
			| match "^robot:admin2$"
		`).Run())

	// Now 'admin2' is the only admin. Login as admin2, and swap 'admin' back in
	// (so that deactivateAuth() runs), and call 'list-admin' (to make sure it
	// works for non-admins)
	loginAsUser(t, "robot:admin2")
	require.NoError(t, tu.BashCmd(`
		pachctl auth modify-admins --add robot:admin --remove robot:admin2
	`).Run())
	require.NoError(t, backoff.Retry(func() error {
		return tu.BashCmd(`pachctl auth list-admins \
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

// TestGetAuthTokenNoSubject tests that 'pachctl get-auth-token' infers the
// subject from the currently logged-in user if none is specified on the command
// line
func TestGetAuthTokenNoSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	alice := tu.UniqueString("robot:alice")
	loginAsUser(t, alice)
	require.NoError(t, tu.BashCmd(`pachctl auth get-auth-token -q | pachctl auth use-auth-token
		pachctl auth whoami | match {{.alice}}
		`,
		"alice", alice,
	).Run())
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
	loginAsUser(t, alice)

	var tokenBuf bytes.Buffer
	tokenCmd := tu.BashCmd(`pachctl auth get-auth-token --ttl=5s -q`)
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

// TestGetOneTimePasswordNoSubject tests that 'pachctl get-otp' infers the
// subject from the currently logged-in user if none is specified on the command
// line
func TestGetOneTimePasswordNoSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	alice := tu.UniqueString("robot:alice")
	loginAsUser(t, alice)

	require.NoError(t, tu.BashCmd(`
		otp="$(pachctl auth get-otp)"
		echo "${otp}" | pachctl auth login --one-time-password
		pachctl auth whoami | match {{.alice}}
		`,
		"alice", alice,
	).Run())
}

// TestGetOneTimePasswordTTL tests that the --ttl argument to 'pachctl get-otp'
// correctly limits the lifetime of the returned token
func TestGetOneTimePasswordTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.ActivateAuth(t)
	defer tu.DeleteAll(t)

	alice := tu.UniqueString("robot:alice")
	loginAsUser(t, alice)

	var otpBuf bytes.Buffer
	otpCmd := tu.BashCmd(`pachctl auth get-otp --ttl=5s`)
	otpCmd.Stdout = &otpBuf
	require.NoError(t, otpCmd.Run())
	otp := strings.TrimSpace(otpBuf.String())

	// wait for OTP to expire
	time.Sleep(6 * time.Second)
	var errMsg bytes.Buffer
	login := tu.BashCmd(`
		echo {{.otp}} | pachctl auth login --one-time-password
	`, "otp", otp)
	login.Stderr = &errMsg
	require.YesError(t, login.Run())
	require.Matches(t, "otp is invalid or has expired", errMsg.String())
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
