package cmds

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
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
	if !strings.Contains(string(out), "ACTIVE") {
		// Enterprise not active in the cluster. Activate it.
		cmd := tu.Cmd("pachctl", "enterprise", "activate")
		cmd.Stdin = strings.NewReader(fmt.Sprintf("%s\n", tu.GetTestEnterpriseCode(t)))
		require.NoError(t, cmd.Run())
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

func TestConnectorCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	require.NoError(t, tu.BashCmd(`
		echo '{}' | pachctl idp create connector --id {{.id}} --name 'testconn' --type 'github' --config -
		pachctl idp list connector | match '{{.id}}'
		pachctl idp get connector {{.id}} \
		  | match 'name: testconn' \
		  | match 'type: github' \
		  | match 'version: 0' \
		  | match "{}" 
		echo '{"client_id": "a"}' | pachctl idp update connector {{.id}} --version 1 --name 'newname' --config -
		pachctl idp get connector {{.id}} \
		  | match 'name: newname' \
		  | match 'type: github' \
		  | match 'version: 1' \
		  | match '{"client_id": "a"}'
		pachctl idp update connector {{.id}} --version 2 --name 'newname2'
		pachctl idp get connector {{.id}} \
		  | match 'name: newname2' \
		  | match 'type: github' \
		  | match 'version: 2' \
		  | match '{"client_id": "a"}'
		pachctl idp delete connector {{.id}}
		`,
		"id", tu.UniqueString("connector"),
	).Run())
}

func TestClientCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	require.NoError(t, tu.BashCmd(`
		pachctl idp create client --id {{.id}} --name 'testclient' --secret 'a secret' --redirectUris https://localhost:1234 \
		  | match 'secret: "a secret"'
		pachctl idp list client | match '{{.id}}'
		pachctl idp get client {{.id}} \
		  | match 'name: testclient' \
		  | match 'secret: a secret' \
		  | match 'redirect URIs: https://localhost:1234' \
		  | match 'trusted peers: ' 
		pachctl idp update client {{.id}} --name 'newname' --redirectUris https://localhost:1234,https://localhost:5678 --trustedPeers x,y,z
		pachctl idp get client {{.id}} \
		  | match 'name: newname' \
		  | match 'secret: a secret' \
		  | match 'redirect URIs: https://localhost:1234, https://localhost:5678' \
		  | match 'trusted peers: x, y, z' 
		pachctl idp update client {{.id}} --name 'newname2'
		pachctl idp get client {{.id}} \
	  	  | match 'name: newname2' \
		  | match 'secret: a secret' \
		  | match 'redirect URIs: https://localhost:1234, https://localhost:5678' \
		  | match 'trusted peers: x, y, z' 
		pachctl idp delete client {{.id}}
		`,
		"id", tu.UniqueString("client"),
	).Run())
}

func TestGetSetConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	activateAuth(t)
	defer deactivateAuth(t)
	require.NoError(t, tu.BashCmd(`
		pachctl idp set config --issuer 'http://example.com:1234'
		pachctl idp get config | match 'issuer: "http://example.com:1234"' 
		`,
		"id", tu.UniqueString("connector"),
	).Run())
}
