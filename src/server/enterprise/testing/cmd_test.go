//go:build k8s

// testing contains integration tests which run against two servers: a pachd, and an enterprise server.
// By contrast, the tests in the server package run against a single pachd.
package testing

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

const enterpriseRootToken = "iamenterprise"

func resetClusterState(ctx context.Context, t *testing.T, c *client.APIClient) {
	ec, err := client.NewEnterpriseClientForTest(ctx)
	require.NoError(t, err)
	// Set the root token, in case a previous test failed
	ec.SetAuthToken(enterpriseRootToken)
	require.NoError(t, ec.DeleteAllEnterprise())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
	pachctl config set context  --overwrite enterprise <<EOF
	{
	  "source": 1,
	  "pachd_address": "grpc://{{ .host }}:{{ .port }}",
	  "session_token": "{{ .token }}"
	}
	EOF`,
		"host", ec.GetAddress().Host,
		"port", fmt.Sprint(ec.GetAddress().Port),
		"token", enterpriseRootToken,
	).Run())
	require.NoError(t, tu.PachctlBashCmd(t, c, "pachctl config set active-enterprise-context enterprise").Run())
}

// TestRegisterPachd tests registering a pachd with the enterprise server when auth is disabled
func TestRegisterPachd(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{.pach_address}}
		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
		"pach_address", pachAddress,
	).Run())
}

// TestRegisterAuthenticated tests registering a pachd with the enterprise server when auth is enabled
func TestRegisterAuthenticated(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	cluster := tu.UniqueString("cluster")
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}

		pachctl enterprise get-state | match ACTIVE
		pachctl license list-clusters \
		  | match 'id: {{.id}}' \
		  | match -v 'last_heartbeat: <nil>'

		pachctl auth whoami --enterprise | match 'pach:root'
	`,
		"id", cluster,
		"license", tu.GetTestEnterpriseCode(t),
		"enterprise_token", enterpriseRootToken,
		"pach_address", pachAddress,
	).Run())
}

// TestEnterpriseRoleBindings tests configuring role bindings for the enterprise server
func TestEnterpriseRoleBindings(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		pachctl auth set enterprise clusterAdmin robot:test1
		pachctl auth get enterprise | match robot:test1
		pachctl auth get cluster | match -v robot:test1
		`,
		"id", tu.UniqueString("cluster"),
		"license", tu.GetTestEnterpriseCode(t),
		"enterprise_token", enterpriseRootToken,
		"token", tu.RootToken,
		"pach_address", pachAddress,
	).Run())
}

// TestGetAndUseRobotToken tests getting a robot token for the enterprise server
func TestGetAndUseRobotToken(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}
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
		"pach_address", pachAddress,
	).Run())
}

// TestConfig tests getting and setting OIDC configuration for the identity server
func TestConfig(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
			`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"pach_address", pachAddress,
	).Run())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth set-config --enterprise <<EOF
{
	"issuer": "http://pach-enterprise.enterprise:31658",
        "localhost_issuer": true,
	"client_id": localhost,
	"redirect_uri": "http://pach-enterprise.enterprise:31650"
}
EOF
	`).Run())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
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
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	ec, err := client.NewEnterpriseClientForTest(ctx)
	require.NoError(t, err)
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		echo '{"id": "test", "name": "test", "type": "mockPassword", "config": {"username": "admin", "password": "password"}}' | pachctl idp create-connector
		`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"pach_address", pachAddress,
	).Run())

	cmd := tu.PachctlBashCmd(t, c, "pachctl auth login --no-browser --enterprise")
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
	require.NoError(t, cmd.Wait())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami --enterprise | match user:{{.user}}
		pachctl auth whoami | match pach:root`,
		"user", tu.DexMockConnectorEmail,
	).Run())
}

// TestLoginPachd tests logging in to pachd
func TestLoginPachd(t *testing.T) {
	ctx := pctx.TestContext(t)
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)

	ec, err := client.NewEnterpriseClientForTest(ctx)
	require.NoError(t, err)
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}
		echo {{.token}} | pachctl auth activate --supply-root-token --client-id pachd2
		echo '{"id": "test", "name": "test", "type": "mockPassword", "config": {"username": "admin", "password": "password"}}' | pachctl idp create-connector
		`,
		"id", tu.UniqueString("cluster"),
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"pach_address", pachAddress,
	).Run())

	cmd := tu.PachctlBashCmd(t, c, "pachctl auth login --no-browser")
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
	require.NoError(t, cmd.Wait())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl auth whoami | match user:{{.user}}
		pachctl auth whoami --enterprise | match 'pach:root'`,
		"user", tu.DexMockConnectorEmail,
	).Run())
}

// Tests synching contexts from the enterprise server
func TestSyncContexts(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	id := tu.UniqueString("cluster")
	clusterId := tu.UniqueString("clusterDeploymentId")
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	// register a new cluster
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }} --pachd-user-address grpc://pachd.default:1655 --cluster-deployment-id {{.clusterId}}
		`,
		"id", id,
		"token", tu.RootToken,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", clusterId,
		"pach_address", pachAddress,
	).Run())

	// assert the registered cluster isn't reflected in the user's config
	require.YesError(t, tu.PachctlBashCmd(t, c, `
		pachctl config list context | match {{.id}}
		`,
		"id", id,
	).Run())

	// sync contexts and assert that the newly registered cluster is accessible
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl enterprise sync-contexts
		pachctl config list context | match {{.id}}
		pachctl config get context {{.id}} | match "\"pachd_address\":[[:space:]]+\"grpc://pachd.default:1655\""
		pachctl config get context {{.id}} | match "\"cluster_deployment_id\":[[:space:]]+\"{{.clusterId}}\""
		pachctl config get context {{.id}} | match "\"source\":[[:space:]]+\"IMPORTED\","
		`,
		"id", id,
		"clusterId", clusterId,
	).Run())

	// re-register cluster with the same cluster ID and new user address
	// the user-address should be updated on sync
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl license update-cluster --id {{.id}} --user-address {{.userAddress}}
		pachctl enterprise sync-contexts
		pachctl config get context {{.id}} | match "\"pachd_address\":[[:space:]]+\"{{.userAddress}}\""
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
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl license update-cluster --id {{.id}} --cluster-deployment-id {{.clusterId}}
		pachctl enterprise sync-contexts
		pachctl config get context {{.id}} | match "\"pachd_address\":[[:space:]]+\"{{.userAddress}}\""
		pachctl config get context {{.id}} | match "\"cluster_deployment_id\":[[:space:]]+\"{{.clusterId}}\""
		`,
		"id", id,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", newClusterId,
		"userAddress", "grpc://pachd.default:700",
	).Run())

	// make sure that the cluster with id = 'localhost' does not get synched, which is
	// self referencing context record for the enterprise server.
	// it should be filtered on the criteria of being set as an enterprise server record
	require.YesError(t, tu.PachctlBashCmd(t, c, `pachctl config list context | match localhost`).Run())
}

// Tests RegisterCluster command's derived argument values if not provided
func TestRegisterDefaultArgs(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)

	id := tu.UniqueString("cluster")

	// get cluster ID from connection
	clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &admin.InspectClusterRequest{
		ClientVersion: version.Version,
	})
	require.NoError(t, inspectErr)
	clusterId := clusterInfo.DeploymentId

	host := c.GetAddress().Host
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	// register a new cluster
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		echo {{.enterprise_token}} | pachctl auth activate --enterprise --issuer http://pach-enterprise.enterprise:31658 --supply-root-token
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address}}

		pachctl enterprise sync-contexts

		pachctl config list context | match {{.id}}
		pachctl config get context {{.id}} | match "\"pachd_address\":[[:space:]]+\"{{.list_pach_address}}"
		pachctl config get context {{.id}} | match "\"cluster_deployment_id\":[[:space:]]+\"{{.clusterId}}\""
		pachctl config get context {{.id}} | match "\"source\":[[:space:]]+\"IMPORTED\","
		`,
		"id", id,
		"enterprise_token", enterpriseRootToken,
		"license", tu.GetTestEnterpriseCode(t),
		"clusterId", clusterId,
		"pach_address", pachAddress,
		"list_pach_address", fmt.Sprintf("grpc://%s:%v", host, c.GetAddress().Port), // assert that a localhost address is registered
	).Run())
}

// tests that Cluster Registration is undone when enterprise service fails to activate in the `enterprise register` subcommand
func TestRegisterRollback(t *testing.T) {
	ctx := context.Background()
	c, ns := minikubetestenv.AcquireCluster(t, minikubetestenv.EnterpriseMemberOption)
	resetClusterState(ctx, t, c)
	defer resetClusterState(ctx, t, c)
	id := tu.UniqueString("cluster")

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		echo {{.license}} | pachctl license activate
		`,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())
	pachAddress := fmt.Sprintf("grpc://pachd.%s:%v", ns, c.GetAddress().Port)
	// passing an unreachable enterprise-server-address to the `enterprise register` command
	// causes it to fail, and rollback cluster record creation
	require.YesError(t, tu.PachctlBashCmd(t, c, `
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc:/bad-address:31650 --pachd-address {{ .pach_address }}
		`,
		"id", id,
		"pach_address", pachAddress,
	).Run())

	// verify the cluster id is not present in the license server's registered clusters
	require.YesError(t, tu.PachctlBashCmd(t, c, `
		pachctl license list-clusters \
			| match 'id: {{.id}}' \
		`,
		"id", id,
	).Run())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl enterprise register --id {{.id}} --enterprise-server-address grpc://pach-enterprise.enterprise:31650 --pachd-address {{ .pach_address }}
		`,
		"id", id,
		"pach_address", pachAddress,
	).Run())
}
