package cmds

// Tests to add:
// basic login test
// login with no auth service deployed

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var activateMut sync.Mutex

func activateAuth(t *testing.T) {
	t.Helper()
	activateMut.Lock()
	defer activateMut.Unlock()
	// TODO(msteffen): Make sure client & server have the same version

	// Check if Pachyderm Enterprise is active -- if not, activate it
	cmd := tu.Cmd("pachctl", "enterprise", "get-state")
	out, err := cmd.Output()
	require.NoError(t, err)
	if string(out) != "ACTIVE" {
		cmd = tu.Cmd("pachctl", "enterprise", "activate", tu.GetTestEnterpriseCode())
		// run cmd in retry loop--sometimes s3 reads flake
		require.NoError(t, backoff.Retry(func() error {
			return cmd.Run()
		}, backoff.NewTestingBackOff()))
	}

	// Logout (to clear any expired tokens) and activate Pachyderm auth
	require.NoError(t, tu.Cmd("pachctl", "auth", "logout").Run())
	cmd = tu.Cmd("pachctl", "auth", "activate")
	cmd.Stdin = strings.NewReader("admin\n")
	require.NoError(t, cmd.Run())
}

func deactivateAuth(t *testing.T) {
	t.Helper()
	activateMut.Lock()
	defer activateMut.Unlock()

	// Check if Pachyderm Auth is active -- if so, deactivate it
	if err := tu.BashCmd("echo admin | pachctl auth login").Run(); err == nil {
		require.NoError(t, tu.BashCmd("yes | pachctl auth deactivate").Run())
	}
}

func TestAuthBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	require.NoError(t, tu.BashCmd(`
		echo "{{.alice}}" | pachctl auth login
		pachctl create-repo {{.repo}}
		pachctl list-repo \
			| match {{.repo}}
		pachctl inspect-repo {{.repo}}
		`,
		"alice", tu.UniqueString("alice"),
		"repo", tu.UniqueString("TestAuthBasic-repo"),
	).Run())
}

func TestWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	require.NoError(t, tu.BashCmd(`
		echo "{{.alice}}" | pachctl auth login
		pachctl auth whoami | match {{.alice}}
		`,
		"alice", tu.UniqueString("alice"),
	).Run())
}

func TestCheckGetSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	// Test both forms of the 'pachctl auth get' command, as well as 'pachctl auth check'
	require.NoError(t, tu.BashCmd(`
		echo "{{.alice}}" | pachctl auth login
		pachctl create-repo {{.repo}}
		pachctl auth check owner {{.repo}}
		pachctl auth get {{.repo}} \
			| match {{.alice}}
		pachctl auth get {{.bob}} {{.repo}} \
			| match NONE
		`,
		"alice", tu.UniqueString("alice"),
		"bob", tu.UniqueString("bob"),
		"repo", tu.UniqueString("TestGet-repo"),
	).Run())

	// Test 'pachctl auth set'
	require.NoError(t, tu.BashCmd(`
		echo "{{.alice}}" | pachctl auth login
		pachctl create-repo {{.repo}}
		pachctl auth set {{.bob}} reader {{.repo}}
		pachctl auth get {{.bob}} {{.repo}} \
			| match READER
		`,
		"alice", tu.UniqueString("alice"),
		"bob", tu.UniqueString("bob"),
		"repo", tu.UniqueString("TestGet-repo"),
	).Run())
}

func TestAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)

	// Modify the list of admins to replace 'admin' with 'admin2'
	require.NoError(t, tu.BashCmd("echo admin | pachctl auth login").Run())
	require.NoError(t, tu.BashCmd(`
		pachctl auth list-admins \
			| match "admin"
		pachctl auth modify-admins --add admin2
		pachctl auth list-admins \
			| match  "admin2"
		pachctl auth modify-admins --remove admin

		# as 'admin' is a substr of 'admin2', use '^admin$' regex...
		pachctl auth list-admins \
			| match -v "^github:admin$" \
			| match "^github:admin2$"
		`).Run())

	// Now 'admin2' is the only admin. Login as admin2, and swap 'admin' back in
	// (so that deactivateAuth() runs), and call 'list-admin' (to make sure it
	// works for non-admins)
	require.NoError(t, tu.BashCmd("echo admin2 | pachctl auth login").Run())
	require.NoError(t, tu.BashCmd(`
		pachctl auth modify-admins --add admin --remove admin2
		pachctl auth list-admins \
			| match -v "admin2" \
			| match "admin"
		`).Run())
}

func TestModifyAdminsPropagateError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)

	// Add admin2, and then try to remove it along with a fake admin. Make sure we
	// get an error
	require.NoError(t, tu.BashCmd("echo admin | pachctl auth login").Run())
	require.NoError(t, tu.BashCmd(`
		pachctl auth list-admins \
			| match "admin"
		pachctl auth modify-admins --add admin2
		pachctl auth list-admins \
			| match  "admin2"

 		# cmd should fail
		! pachctl auth modify-admins --remove admin1,not_in_list
		`).Run())
}

func TestMain(m *testing.M) {
	// Preemptively deactivate Pachyderm auth (to avoid errors in early tests)
	if err := tu.BashCmd("echo 'admin' | pachctl auth login").Run(); err == nil {
		if err := tu.BashCmd("yes | pachctl auth deactivate").Run(); err != nil {
			panic(err.Error())
		}
	}
	time.Sleep(5 * time.Second)
	backoff.Retry(func() error {
		cmd := tu.Cmd("pachctl", "auth", "login")
		cmd.Stdin = strings.NewReader("admin\n")
		if cmd.Run() != nil {
			return nil // success -- auth is deactivated
		}
		return errors.New("auth not deactivated yet")
	}, backoff.NewTestingBackOff())
	os.Exit(m.Run())
}
