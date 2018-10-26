package cmds

// Tests to add:
// basic login test
// login with no auth service deployed

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
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

// activateEnterprise checks if Pachyderm Enterprise is active and if not,
// activates it
func activateEnterprise(t *testing.T) {
	cmd := tu.Cmd("pachctl", "enterprise", "get-state")
	out, err := cmd.Output()
	require.NoError(t, err)
	if string(out) != "ACTIVE" {
		// Enterprise not active in the cluster. Activate it
		require.NoError(t,
			tu.Cmd("pachctl", "enterprise", "activate", tu.GetTestEnterpriseCode()).Run())
	}

}

func activateAuth(t *testing.T) {
	t.Helper()
	activateMut.Lock()
	defer activateMut.Unlock()
	activateEnterprise(t)
	// TODO(msteffen): Make sure client & server have the same version
	// Logout (to clear any expired tokens) and activate Pachyderm auth
	require.NoError(t, tu.Cmd("pachctl", "auth", "logout").Run())
	cmd := tu.Cmd("pachctl", "auth", "activate")
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

	// Wait for auth to finish deactivating
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

func TestGetAndUseAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)

	// Test both get-auth-token and use-auth-token; make sure that they work
	// together with -q
	require.NoError(t, tu.BashCmd("echo admin | pachctl auth login").Run())
	require.NoError(t, tu.BashCmd(`
	pachctl auth get-auth-token -q robot:marvin \
	  | pachctl auth use-auth-token
	pachctl auth whoami \
	  | match 'robot:marvin'
		`).Run())
}

func TestActivateAsRobotUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// We need a custom 'activate' command, so reproduce 'activateAuth' minus the
	// actual call
	defer deactivateAuth(t) // unwind "activate" command before deactivating
	activateMut.Lock()
	defer activateMut.Unlock()
	activateEnterprise(t)
	// Logout (to clear any expired tokens) and activate Pachyderm auth
	require.NoError(t, tu.BashCmd(`
	pachctl auth logout
	pachctl auth activate --initial-admin=robot:hal9000
	pachctl auth whoami \
		| match 'robot:hal9000'
	`).Run())

	// Make "admin" a cluster admins, so that deactivateAuth works
	require.NoError(t,
		tu.Cmd("pachctl", "auth", "modify-admins", "--add=admin").Run())
}

func TestActivateMismatchedUsernames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// We need a custom 'activate' command, to reproduce 'activateAuth' minus the
	// actual call
	activateMut.Lock()
	defer activateMut.Unlock()
	activateEnterprise(t)
	// Logout (to clear any expired tokens) and activate Pachyderm auth
	activate := tu.BashCmd(`
		pachctl auth logout
		echo alice | pachctl auth activate --initial-admin=bob
	`)
	var errorMsg bytes.Buffer
	activate.Stderr = &errorMsg
	require.YesError(t, activate.Run())
	require.Matches(t, "github:alice", errorMsg.String())
	require.Matches(t, "github:bob", errorMsg.String())
}

func TestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)

	idpMetadata := base64.StdEncoding.EncodeToString([]byte(`<EntityDescriptor
		  xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
		  validUntil="` + time.Now().Format(time.RFC3339) + `"
		  entityID="metadata">
      <SPSSODescriptor
		    xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
		    validUntil="` + time.Now().Format(time.RFC3339) + `"
		    protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"
		    AuthnRequestsSigned="false"
		    WantAssertionsSigned="true">
        <AssertionConsumerService
		      Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
		      Location="acs"
		      index="1">
		    </AssertionConsumerService>
      </SPSSODescriptor>
    </EntityDescriptor>`))
	require.NoError(t, tu.BashCmd(`
		echo "admin" | pachctl auth login
		pachctl auth set-config <<EOF
		{
		  "live_config_version": 0,
		  "id_providers": [{
		    "name": "idp",
		    "description": "fake ID provider for testing",
		    "saml": {
		      "metadata_xml": "`+idpMetadata+`"
		    }
		  }],
		  "saml_svc_options": {
		    "acs_url": "http://www.example.com",
		    "metadata_url": "http://www.example.com"
		  }
		}
		EOF
		pachctl auth get-config \
		  | match '"live_config_version": 1,' \
		  | match '"saml_svc_options": {' \
		  | match '"acs_url": "http://www.example.com",' \
		  | match '"metadata_url": "http://www.example.com"' \
		  | match '}'
		`).Run())
}

func TestMain(m *testing.M) {
	// Preemptively deactivate Pachyderm auth (to avoid errors in early tests)
	if err := tu.BashCmd("echo 'admin' | pachctl auth login &>/dev/null").Run(); err == nil {
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
