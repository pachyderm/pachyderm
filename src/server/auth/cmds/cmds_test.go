package cmds

// Tests to add:
// basic login test
// login with no auth service deployed

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}

// C is a convenience function that replaces exec.Command. It's both shorter
// and it uses the current process's stderr as output for the command, which
// makes debugging failures much easier (i.e. you get an error message
// rather than "exit status 1")
func C(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	// for convenience, simulate hitting "enter" after any prompt. This can easily
	// be replaced
	cmd.Stdin = strings.NewReader("\n")
	return cmd
}

// BashC is a convenience function that replaces exec.Command, running the given
// string in a bash shell
func BashC(cmds string) *exec.Cmd {
	cmd := exec.Command("bash", "-c", cmds)
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader("\n")
	return cmd
}

// sub is a convenience function to do inline template substitution
func sub(s string, subs ...string) string {
	buf := &bytes.Buffer{}
	if len(subs)%2 == 1 {
		panic("some variable does not have a corresponding value")
	}

	// copy 'subs' into a map
	subsMap := make(map[string]string)
	for i := 0; i < len(subs); i += 2 {
		subsMap[subs[i]] = subs[i+1]
	}

	// do the substitution
	template.Must(template.New("").Parse(s)).Execute(buf, subsMap)
	return buf.String()
}

var activateMut sync.Mutex

func activateAuth(t *testing.T) {
	activateMut.Lock()
	defer activateMut.Unlock()
	// TODO(msteffen): Make sure client & server have the same version

	// Check if Pachyderm Enterprise is active -- if not, activate it
	cmd := C("pachctl", "enterprise", "get-state")
	out, err := cmd.Output()
	require.NoError(t, err)
	if string(out) != "ACTIVE" {
		cmd = C("pachctl", "enterprise", "activate", testutil.GetTestEnterpriseCode())
		require.NoError(t, cmd.Run())
	}

	// Logout (to clear any expired tokens) and activate Pachyderm auth
	C("pachctl", "auth", "logout").Run()
	C("pachctl", "auth", "activate", "--user=admin").Run()
}

func deactivateAuth(t *testing.T) {
	activateMut.Lock()
	defer activateMut.Unlock()

	// Check if Pachyderm Auth is active -- if so, deactivate it
	cmd := C("pachctl", "auth", "login", "-u", "admin")
	require.NoError(t, cmd.Run())
	cmd = BashC("yes | pachctl auth deactivate")
	require.NoError(t, cmd.Run())
}

func TestAuthBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	cmd := BashC(sub(`
		echo "\n" | pachctl auth login -u {{.alice}}
		pachctl create-repo {{.repo}}
		pachctl list-repo \
			| grep -q {{.repo}}
		pachctl inspect-repo {{.repo}}
		`,
		"alice", uniqueString("alice"),
		"repo", uniqueString("TestAuthBasic-repo"),
	))
	require.NoError(t, cmd.Run())
}

func TestWhoAmI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	cmd := BashC(sub(`
		pachctl auth login -u {{.alice}}
		pachctl auth whoami | grep -q {{.alice}}
		`,
		"alice", uniqueString("alice"),
	))
	require.NoError(t, cmd.Run())
}

func TestCheckGetSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	// Test both forms of the 'pachctl auth get' command, as well as 'pachctl auth check'
	cmd := BashC(sub(`
		pachctl auth login -u {{.alice}}
		pachctl create-repo {{.repo}}
		pachctl auth check owner {{.repo}}
		pachctl auth get {{.repo}} \
			| grep -q {{.alice}}
		pachctl auth get {{.bob}} {{.repo}} \
			| grep -q NONE
		`,
		"alice", uniqueString("alice"),
		"bob", uniqueString("bob"),
		"repo", uniqueString("TestGet-repo"),
	))
	require.NoError(t, cmd.Run())

	// Test 'pachctl auth set'
	cmd = BashC(sub(`
		pachctl auth login -u {{.alice}}
		pachctl create-repo {{.repo}}
		pachctl auth set {{.bob}} reader {{.repo}}
		pachctl auth get {{.bob}} {{.repo}} \
			| grep -q READER
		`,
		"alice", uniqueString("alice"),
		"bob", uniqueString("bob"),
		"repo", uniqueString("TestGet-repo"),
	))
	require.NoError(t, cmd.Run())
}

func TestAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)

	// Modify the list of admins to replace 'admin' with 'admin2'
	require.NoError(t, C("pachctl", "auth", "login", "-u", "admin").Run())
	cmd := BashC(sub(`
		pachctl auth list-admins \
			| grep -q "admin"
		pachctl auth modify-admins --add admin2
		pachctl auth list-admins \
			| grep -q  "admin2"
		pachctl auth modify-admins --remove admin

		# as 'admin' is a substr of 'admin2', use '^admin$' regex...
		pachctl auth list-admins \
			| grep -v "^admin$" \
			| grep -q "^admin2$"
		`))
	require.NoError(t, cmd.Run())

	// Now 'admin2' is the only admin. Login as admin2, and swap 'admin' back in
	// (so that deactivateAuth() runs), and call 'list-admin' (to make sure it
	// works for non-admins)
	require.NoError(t, C("pachctl", "auth", "login", "-u", "admin2").Run())
	cmd = BashC(sub(`
		pachctl auth modify-admins --add admin --remove admin2
		pachctl auth list-admins \
			| grep -v "admin2" \
			| grep -q "admin"
		`,
		"alice", uniqueString("alice"),
		"bob", uniqueString("bob"),
	))
	require.NoError(t, cmd.Run())
}

func TestMain(m *testing.M) {
	// Preemptively deactivate Pachyderm auth (to avoid errors in early tests)
	cmd := C("pachctl", "auth", "login", "-u", "admin")
	cmd.Run()
	cmd = BashC("yes | pachctl auth deactivate")
	cmd.Run()
	time.Sleep(5 * time.Second)
	backoff.Retry(func() error {
		cmd := C("pachctl", "auth", "login", "-u", "admin")
		if cmd.Run() != nil {
			return nil // success -- auth is deactivated
		}
		return errors.New("auth not deactivated yet")
	}, backoff.NewTestingBackOff())
	os.Exit(m.Run())
}
