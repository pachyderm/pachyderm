//go:build !k8s

package server_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

// TestActivate tests that we can activate the license server
// by providing a valid enterprise activation code. This is exercised
// in a bunch of other tests, but in the interest of being explicit
// this test only focuses on activation.
func TestActivate(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))

	// Activate Enterprise
	tu.ActivateEnterprise(t, client, peerPort)

	// Confirm we can get the activation code back
	resp, err := client.License.GetActivationCode(client.Ctx(), &license.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_ACTIVE, resp.State)
	require.Equal(t, tu.GetTestEnterpriseCode(t), resp.ActivationCode)
}

// TestExpired tests that the license server returns the expired state
// if the expiration of the license is in the past.
func TestExpired(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateEnterprise(t, client, peerPort)

	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)

	// Activate Enterprise with an expiration in the past
	_, err = client.License.Activate(context.Background(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)

	// Confirm the license server state is expired
	resp, err := client.License.GetActivationCode(client.Ctx(), &license.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_EXPIRED, resp.State)
	require.Equal(t, tu.GetTestEnterpriseCode(t), resp.ActivationCode)
}

// TestGetActivationCodeNotAdmin tests that non-admin users cannot retrieve
// the enterprise activation code
func TestGetActivationCodeNotAdmin(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateAuthClient(t, client, peerPort)
	aliceClient := tu.AuthenticateClient(t, client, "robot:alice")
	_, err := aliceClient.License.GetActivationCode(aliceClient.Ctx(), &license.GetActivationCodeRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

// TestDeleteAll tests that DeleteAll removes all registered clusters and
// puts the license server in the NONE state.
func TestDeleteAll(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateEnterprise(t, client, peerPort)

	// Confirm one cluster is registered
	clusters, err := client.License.ListClusters(client.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(clusters.Clusters))

	// Call DeleteAll
	_, err = client.License.DeleteAll(client.Ctx(), &license.DeleteAllRequest{})
	require.NoError(t, err)

	// No license is registered
	resp, err := client.License.GetActivationCode(client.Ctx(), &license.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_NONE, resp.State)

	// Activate Enterprise but don't register any clusters
	_, err = client.License.Activate(context.Background(),
		&license.ActivateRequest{
			ActivationCode: tu.GetTestEnterpriseCode(t),
		})
	require.NoError(t, err)

	// No clusters are registered
	clusters, err = client.License.ListClusters(client.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(clusters.Clusters))
}

// TestDeleteAllNotAdmin confirms only admins can call DeleteAll when auth is enabled
func TestDeleteAllNotAdmin(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateAuthClient(t, c, peerPort)
	aliceClient := tu.AuthenticateClient(t, c, "robot:alice")
	_, err := aliceClient.License.DeleteAll(aliceClient.Ctx(), &license.DeleteAllRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

// TestClusterCRUD tests that clusters can be added, listed, updated and deleted
func TestClusterCRUD(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateAuthClient(t, client, peerPort)

	clusters, err := client.License.ListClusters(client.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(clusters.Clusters))

	// Add a new cluster
	newCluster, err := client.License.AddCluster(client.Ctx(), &license.AddClusterRequest{
		Id:                  "new",
		Address:             "grpc://localhost:" + peerPort,
		UserAddress:         "grpc://localhost:999",
		ClusterDeploymentId: "some-deployment-id",
		EnterpriseServer:    false,
	})
	require.NoError(t, err)
	require.True(t, len(newCluster.Secret) >= 30)

	// Confirm there are now two clusters
	expectedStatuses := map[string]*license.ClusterStatus{
		"localhost": {
			Id:      "localhost",
			Address: "grpc://localhost:" + peerPort,
		},
		"new": {
			Id:      "new",
			Address: "grpc://localhost:" + peerPort,
		},
	}

	expectedUserClusters := map[string]*license.UserClusterInfo{
		"new": {
			Id:                  "new",
			Address:             "grpc://localhost:999",
			ClusterDeploymentId: "some-deployment-id",
			EnterpriseServer:    false,
		},
	}

	verifyListClustersContents := func(expected map[string]*license.ClusterStatus, actual []*license.ClusterStatus) {
		for _, v := range actual {
			require.Equal(t, expected[v.Id].Address, v.Address)
			require.Equal(t, expected[v.Id].AuthEnabled, v.AuthEnabled)
		}
	}

	verifyListUserClustersContents := func(expected map[string]*license.UserClusterInfo, actual []*license.UserClusterInfo) {
		for _, v := range actual {
			require.Equal(t, expected[v.Id].Address, v.Address)
			require.Equal(t, expected[v.Id].ClusterDeploymentId, v.ClusterDeploymentId)
			require.Equal(t, expected[v.Id].EnterpriseServer, v.EnterpriseServer)
		}
	}

	clusters, err = client.License.ListClusters(client.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(clusters.Clusters))
	verifyListClustersContents(expectedStatuses, clusters.Clusters)

	var userClusters *license.ListUserClustersResponse
	userClusters, err = client.License.ListUserClusters(client.Ctx(), &license.ListUserClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(userClusters.Clusters))
	verifyListUserClustersContents(expectedUserClusters, userClusters.Clusters)

	// Update the cluster
	_, err = client.License.UpdateCluster(client.Ctx(), &license.UpdateClusterRequest{
		Id:                  "new",
		Address:             "localhost:1653",
		UserAddress:         "localhost:1000",
		ClusterDeploymentId: "another-deployment-id",
	})
	require.NoError(t, err)

	expectedStatuses["new"].Address = "localhost:1653"

	expectedUserClusters["new"].Address = "localhost:1000"
	expectedUserClusters["new"].ClusterDeploymentId = "another-deployment-id"

	clusters, err = client.License.ListClusters(client.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(clusters.Clusters))
	verifyListClustersContents(expectedStatuses, clusters.Clusters)

	userClusters, err = client.License.ListUserClusters(client.Ctx(), &license.ListUserClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(userClusters.Clusters))
	verifyListUserClustersContents(expectedUserClusters, userClusters.Clusters)

	// Delete the new cluster
	_, err = client.License.DeleteCluster(client.Ctx(), &license.DeleteClusterRequest{
		Id: "new",
	})
	require.NoError(t, err)

	clusters, err = client.License.ListClusters(client.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(clusters.Clusters))

	delete(expectedStatuses, "new")

	verifyListClustersContents(expectedStatuses, clusters.Clusters)

	userClusters, err = client.License.ListUserClusters(client.Ctx(), &license.ListUserClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(userClusters.Clusters))
	verifyListUserClustersContents(expectedUserClusters, userClusters.Clusters)
}

// TestAddClusterUnreachable tries to add a cluster with a misconfigured address
// and confirms there's an error
func TestAddClusterAddressValidation(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateEnterprise(t, client, peerPort)

	_, err := client.License.AddCluster(client.Ctx(), &license.AddClusterRequest{
		Id:      "new",
		Address: "",
	})
	require.YesError(t, err)
	require.Matches(t, "no address provided for cluster", err.Error())
}

// TestUpdateClusterUnreachable tries to update an existing cluster with a misconfigured address
// and confirms there's an error
func TestUpdateClusterAddressValidation(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateEnterprise(t, client, peerPort)

	_, err := client.License.UpdateCluster(client.Ctx(), &license.UpdateClusterRequest{
		Id:      "localhost",
		Address: "",
	})
	require.YesError(t, err)
	require.Matches(t, "No cluster fields were provided to the UpdateCluster RPC", err.Error())
}

// TestAddClusterNoLicense tries to add a cluster with no license configured and
// confirms there's an error
func TestAddClusterNoLicense(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	_, err := client.License.AddCluster(client.Ctx(), &license.AddClusterRequest{
		Id:      "new",
		Address: "grpc://localhost:" + peerPort,
	})
	require.YesError(t, err)
	require.Matches(t, "enterprise license is not valid", err.Error())
}

// TestClusterCRUDNotAdmin confirms that AddCluster, ListClusters, DeleteCluster and
// UpdateCluster require admin access when auth is enabled
func TestClusterCRUDNotAdmin(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	client := env.PachClient
	tu.ActivateAuthClient(t, client, peerPort)
	aliceClient := tu.AuthenticateClient(t, client, "robot:alice")

	_, err := aliceClient.License.AddCluster(aliceClient.Ctx(), &license.AddClusterRequest{
		Id: "localhost",
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	_, err = aliceClient.License.ListClusters(aliceClient.Ctx(), &license.ListClustersRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	_, err = aliceClient.License.DeleteCluster(aliceClient.Ctx(), &license.DeleteClusterRequest{
		Id: "localhost",
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())

	_, err = aliceClient.License.UpdateCluster(aliceClient.Ctx(), &license.UpdateClusterRequest{
		Id: "localhost",
	})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

// TestHeartbeat adds a new cluster and confirms that the cluster metadata is updated when that cluster
// heartbeats.
func TestHeartbeat(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	rootClient := tu.AuthenticateClient(t, c, auth.RootUser)

	// Confirm the localhost cluster is configured as expected
	clusters, err := rootClient.License.ListClusters(rootClient.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(clusters.Clusters))
	require.Equal(t, false, clusters.Clusters[0].AuthEnabled)

	// Heartbeat using the correct shared secret, confirm the activation code is returned
	pachClient := tu.UnauthenticatedPachClient(t, c)
	resp, err := pachClient.License.Heartbeat(pachClient.Ctx(), &license.HeartbeatRequest{
		Id:          "localhost",
		Secret:      "localhost",
		Version:     "some weird version",
		AuthEnabled: true,
	})
	require.NoError(t, err)
	require.Equal(t, tu.GetTestEnterpriseCode(t), resp.License.ActivationCode)

	// List clusters again, auth_enabled, version and last_heartbeat should be updated
	newClusters, err := rootClient.License.ListClusters(rootClient.Ctx(), &license.ListClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(newClusters.Clusters))
	require.Equal(t, true, newClusters.Clusters[0].AuthEnabled)
	require.Equal(t, "some weird version", newClusters.Clusters[0].Version)
	require.True(t, newClusters.Clusters[0].LastHeartbeat.After(*clusters.Clusters[0].LastHeartbeat))
}

// TestHeartbeatWrongSecret tests that Heartbeat doesn't update the record if the shared secret is incorrect
func TestHeartbeatWrongSecret(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	pachClient := env.PachClient
	_, err := pachClient.License.Heartbeat(pachClient.Ctx(), &license.HeartbeatRequest{
		Id:          "localhost",
		Secret:      "wrong secret",
		Version:     "some weird version",
		AuthEnabled: true,
	})
	require.YesError(t, err)
}

func TestListUserClusters(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	client := env.PachClient
	tu.ActivateAuthClient(t, client, peerPort)

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_ACTIVE, resp.State)

	_, err = client.License.AddCluster(client.Ctx(), &license.AddClusterRequest{
		Id:                  "new",
		Address:             "grpc://localhost:1650",
		UserAddress:         "grpc://localhost:999",
		ClusterDeploymentId: "some-deployment-id",
		EnterpriseServer:    false,
	})
	require.NoError(t, err)

	// Make sure that an added cluster shows up in ListUserClusters
	var userClustersResp *license.ListUserClustersResponse
	userClustersResp, err = client.License.ListUserClusters(client.Ctx(), &license.ListUserClustersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(userClustersResp.Clusters))
	require.Equal(t, "new", userClustersResp.Clusters[0].Id)
	require.Equal(t, "grpc://localhost:999", userClustersResp.Clusters[0].Address)
	require.Equal(t, false, userClustersResp.Clusters[0].EnterpriseServer)
}
