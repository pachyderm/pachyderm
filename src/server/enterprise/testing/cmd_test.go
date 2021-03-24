// testing contains integration tests which run against two servers: a pachd, and an enterprise server.
// By contrast, the tests in the server package run against a single pachd.
package testing

import (
	"bufio"
	"os/exec"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func resetClusterState(t *testing.T) {
	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	c, err := client.NewForTest()
	require.NoError(t, err)

	// Set the root token, in case a previous test failed
	c.SetAuthToken(tu.RootToken)
	ec.SetAuthToken(tu.RootToken)

	require.NoError(t, c.DeleteAll())
	require.NoError(t, ec.DeleteAllEnterprise())
}

// TestRegisterPachd tests registering a pachd with the enterprise server when auth is disabled
func TestRegisterPachd(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
	).Run())
}

// TestRegisterAuthenticated tests registering a pachd with the enterprise server when auth is enabled
func TestRegisterAuthenticated(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	cluster := tu.UniqueString("cluster")
	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		echo {{.token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650

		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'

		pachctl auth whoami --enterprise | match 'pach:root'
	`,
		"id", cluster,
		"license", tu.GetTestEnterpriseCode(t),
		"token", tu.RootToken,
	).Run())
}

// TestEnterpriseRoleBindings tests configuring role bindings for the enterprise server
func TestEnterpriseRoleBindings(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		echo {{.token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		pachctl auth set enterprise clusterAdmin robot:test1
		pachctl auth get enterprise | match robot:test1
		pachctl auth get cluster | match -v robot:test1
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
		"token", tu.RootToken,
	).Run())
}

// TestGetAndUseRobotToken tests getting a robot token for the enterprise server
func TestGetAndUseRobotToken(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		echo {{.token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		pachctl auth get-robot-token --enterprise -q {{.alice}} | pachctl auth use-auth-token --enterprise
		pachctl auth get-robot-token -q {{.bob}} | pachctl auth use-auth-token
		pachctl auth whoami --enterprise | match {{.alice}}
		pachctl auth whoami | match {{.bob}}
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
		"token", tu.RootToken,
		"alice", tu.UniqueString("alice"),
		"bob", tu.UniqueString("bob"),
	).Run())
}

// TestConfig tests getting and setting OIDC configuration for the identity server
func TestConfig(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		echo {{.token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
			`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())

	require.NoError(t, tu.BashCmd(`
		pachctl auth set-config --enterprise <<EOF
{
	"issuer": "http://pach-enterprise.enterprise:658",
        "localhost_issuer": true,
	"client_id": localhost,
	"redirect_uri": "http://pach-enterprise.enterprise:650"
}
EOF
	`).Run())

	require.NoError(t, tu.BashCmd(`
		pachctl auth get-config --enterprise \
		  | match '"issuer": "http://pach-enterprise.enterprise:658"' \
		  | match '"localhost_issuer": true' \
		  | match '"client_id": "localhost"' \
		  | match '"redirect_uri": "http://pach-enterprise.enterprise:650"'
		`,
	).Run())
}

// TestLoginEnterprise tests logging in to the enterprise server
func TestLoginEnterprise(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		echo {{.token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		echo '{"username": "admin", "password": "password"}' | pachctl idp create-connector --id test --type mockPassword --name test  --config -
		`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())

	cmd := exec.Command("pachctl", "auth", "login", "--no-browser", "--enterprise")
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		if strings.HasPrefix(strings.TrimSpace(sc.Text()), "http://") {
			tu.DoOAuthExchange(t, ec, ec, sc.Text())
			break
		}
	}
	cmd.Wait()

	require.NoError(t, tu.BashCmd(`
		pachctl auth whoami --enterprise | match user:{{.user}}
		pachctl auth whoami | match pach:root`,
		"user", tu.DexMockConnectorEmail,
	).Run())
}

// TestLoginPachd tests logging in to pachd
func TestLoginPachd(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	c, err := client.NewForTest()
	require.NoError(t, err)

	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl enterprise activate
		echo {{.token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:650 --pachd-address pachd.default:650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		echo '{"username": "admin", "password": "password"}' | pachctl idp create-connector --id test --type mockPassword --name test  --config -
		`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())

	cmd := exec.Command("pachctl", "auth", "login", "--no-browser")
	out, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		if strings.HasPrefix(strings.TrimSpace(sc.Text()), "http://") {
			tu.DoOAuthExchange(t, c, ec, sc.Text())
			break
		}
	}
	cmd.Wait()

	require.NoError(t, tu.BashCmd(`
		pachctl auth whoami | match user:{{.user}}
		pachctl auth whoami --enterprise | match 'pach:root'`,
		"user", tu.DexMockConnectorEmail,
	).Run())
}
