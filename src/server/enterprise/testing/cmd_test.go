// testing contains integration tests which run against two servers: a pachd, and an enterprise server.
// By contrast, the tests in the server package run against a single pachd.
package testing

import (
	"bufio"
	"os/exec"
	"strings"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const enterpriseRootToken = "iamenterprise"

func resetClusterState(t *testing.T) {
	ec, err := client.NewEnterpriseClientForTest()
	require.NoError(t, err)

	c, err := client.NewForTest()
	require.NoError(t, err)

	// Set the root token, in case a previous test failed
	c.SetAuthToken(tu.RootToken)
	ec.SetAuthToken(enterpriseRootToken)

	require.NoError(t, c.DeleteAll())
	require.NoError(t, ec.DeleteAllEnterprise())
}

// TestRegisterPachd tests registering a pachd with the enterprise server when auth is disabled
func TestRegisterPachd(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650
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
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650

		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'

		pachctl auth whoami --enterprise | match 'pach:root'
	`,
		"id", cluster,
		"license", tu.GetTestEnterpriseCode(t),
		"enterprise_token", enterpriseRootToken,
	).Run())
}

// TestEnterpriseRoleBindings tests configuring role bindings for the enterprise server
func TestEnterpriseRoleBindings(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		pachctl auth set enterprise clusterAdmin robot:test1
		pachctl auth get enterprise | match robot:test1
		pachctl auth get cluster | match -v robot:test1
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
		"enterprise_token", enterpriseRootToken,
		"token", tu.RootToken,
	).Run())
}

// TestGetAndUseRobotToken tests getting a robot token for the enterprise server
func TestGetAndUseRobotToken(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		pachctl auth get-robot-token --enterprise -q {{.alice}} | tail -1 | pachctl auth use-auth-token --enterprise
		pachctl auth get-robot-token -q {{.bob}} | pachctl auth use-auth-token
		pachctl auth whoami --enterprise | match {{.alice}}
		pachctl auth whoami | match {{.bob}}
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"alice", tu.UniqueString("alice"),
		"bob", tu.UniqueString("bob"),
	).Run())
}

// TestConfig tests getting and setting OIDC configuration for the identity server
func TestConfig(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:31650 --pachd-address pachd.default:30650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
			`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())

	require.NoError(t, tu.BashCmd(`
		pachctl auth set-config --enterprise <<EOF
{
	"issuer": "http://pach-enterprise.enterprise:31658",
        "localhost_issuer": true,
	"client_id": localhost,
	"redirect_uri": "http://pach-enterprise.enterprise:31650"
}
EOF
	`).Run())

	require.NoError(t, tu.BashCmd(`
		pachctl auth get-config --enterprise \
		  | match '"issuer": "http://pach-enterprise.enterprise:31658"' \
		  | match '"localhost_issuer": true' \
		  | match '"client_id": "localhost"' \
		  | match '"redirect_uri": "http://pach-enterprise.enterprise:31650"'
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
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		echo '{"id": "test", "name": "test", "type": "mockPassword", "config": {"username": "admin", "password": "password"}}' | pachctl idp create-connector
		`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
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
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		echo '{"id": "test", "name": "test", "type": "mockPassword", "config": {"username": "admin", "password": "password"}}' | pachctl idp create-connector
		`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
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

// Tests synching contexts from the enterprise server
func TestSyncContexts(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	id := tu.UniqueString("cluster")
	clusterId := tu.UniqueString("clusterDeploymentId")

	// register a new cluster
	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650 --pachd-user-address grpc://pachd.default:1655 --cluster-deployment-id {{.clusterId}} 
		`,
		"id", id,
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", clusterId,
	).Run())

	// assert the registered cluster isn't reflected in the user's config
	require.YesError(t, tu.BashCmd(`
		pachctl config list context | match {{.id}}
		`,
		"id", id,
	).Run())

	// sync contexts and assert that the newly registered cluster is accessible
	require.NoError(t, tu.BashCmd(`
		pachctl enterprise sync-contexts
		pachctl config list context | match {{.id}}
		pachctl config get context {{.id}} | match "\"pachd_address\": \"grpc://pachd.default:1655\"" 
		pachctl config get context {{.id}} | match "\"cluster_deployment_id\": \"{{.clusterId}}\""
		pachctl config get context {{.id}} | match "\"source\": \"IMPORTED\","
		`,
		"id", id,
		"clusterId", clusterId,
	).Run())

	// re-register cluster with the same cluster ID and new user address
	// the user-address should be updated on sync
	require.NoError(t, tu.BashCmd(`
		pachctl license update-cluster --id {{.id}} --user-address {{.userAddress}}
		pachctl enterprise sync-contexts
		pachctl config get context {{.id}} | match "\"pachd_address\": \"{{.userAddress}}\"" 
		`,
		"id", id,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", clusterId,
		"userAddress", "grpc://pachd.default:700",
	).Run())

	// re-register cluster with a new cluster ID
	// the cluster id should be updated and the session token should be set to empty
	// TODO(acohen4): set session_token so that it can be unset
	newClusterId := tu.UniqueString("clusterDeploymentId")
	require.NoError(t, tu.BashCmd(`
		pachctl license update-cluster --id {{.id}} --cluster-deployment-id {{.clusterId}}
		pachctl enterprise sync-contexts
		pachctl config get context {{.id}} | match "\"pachd_address\": \"{{.userAddress}}\"" 
		pachctl config get context {{.id}} | match "\"cluster_deployment_id\": \"{{.clusterId}}\"" 
		`,
		"id", id,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", newClusterId,
		"userAddress", "grpc://pachd.default:700",
	).Run())

	// make sure that the cluster with id = 'localhost' does not get synched, which is
	// self referencing context record for the enterprise server.
	// it should be filtered on the criteria of being set as an enterprise server record
	require.YesError(t, tu.BashCmd(`pachctl config list context | match localhost`).Run())
}

// Tests RegisterCluster command's derived argument values if not provided
func TestRegisterDefaultArgs(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	c, err := client.NewForTest()
	require.NoError(t, err)

	id := tu.UniqueString("cluster")

	// get cluster ID from connection
	clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &types.Empty{})
	require.NoError(t, inspectErr)
	clusterId := clusterInfo.DeploymentID

	host := c.GetAddress().Host
	if host == "0.0.0.0" {
		host = "localhost"
	}

	// register a new cluster
	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650

		pachctl enterprise sync-contexts

		pachctl config list context | match {{.id}}
		pachctl config get context {{.id}} | match "\"pachd_address\": \"{{.pachdAddress}}"
		pachctl config get context {{.id}} | match "\"cluster_deployment_id\": \"{{.clusterId}}\""
		pachctl config get context {{.id}} | match "\"source\": \"IMPORTED\","
		`,
		"id", id,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", clusterId,
		"pachdAddress", "grpc://"+host+":30650", // assert that a localhost address is registered
	).Run())
}

// tests that Cluster Registration is undone when enterprise service fails to activate in the `enterprise register` subcommand
func TestRegisterRollback(t *testing.T) {
	resetClusterState(t)
	defer resetClusterState(t)

	id := tu.UniqueString("cluster")

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		`,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())

	// passing an unreachable enterprise-server-address to the `enterprise register` command
	// causes it to fail, and rollback cluster record creation
	require.YesError(t, tu.BashCmd(`
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc:/bad-address:31650 --pachd-address grpc://pachd.default:30650
		`,
		"id", id,
	).Run())

	// verify the cluster id is not present in the license server's registered clusters
	require.YesError(t, tu.BashCmd(`
		pachctl license list-clusters \
			| match 'id: {{.id}}' \
		`,
		"id", id,
	).Run())

	require.NoError(t, tu.BashCmd(`
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address grpc://pachd.default:30650
		`,
		"id", id,
	).Run())
}
