package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// TestAuthNotActivated checks that no RPCs can be made when the auth service is disabled
func TestAuthNotActivated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	client := tu.GetPachClient(t)
	_, err := client.SetIdentityConfig(client.Ctx(), &identity.SetIdentityConfigRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.GetIdentityConfig(client.Ctx(), &identity.GetIdentityConfigRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.CreateIDPConnector(client.Ctx(), &identity.CreateIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.GetIDPConnector(client.Ctx(), &identity.GetIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.UpdateIDPConnector(client.Ctx(), &identity.UpdateIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.ListIDPConnectors(client.Ctx(), &identity.ListIDPConnectorsRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.DeleteIDPConnector(client.Ctx(), &identity.DeleteIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.CreateOIDCClient(client.Ctx(), &identity.CreateOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.GetOIDCClient(client.Ctx(), &identity.GetOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.UpdateOIDCClient(client.Ctx(), &identity.UpdateOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.ListOIDCClients(client.Ctx(), &identity.ListOIDCClientsRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.DeleteOIDCClient(client.Ctx(), &identity.DeleteOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.IdentityAPIClient.DeleteAll(client.Ctx(), &identity.DeleteAllRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())
}

// TestUserNotAdmin checks that no RPCs can be made by non-admin users
func TestUserNotAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	alice := tu.UniqueString("alice")
	aliceClient := tu.GetAuthenticatedPachClient(t, alice)

	_, err := aliceClient.SetIdentityConfig(aliceClient.Ctx(), &identity.SetIdentityConfigRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call SetIdentityConfig", alice), err.Error())

	_, err = aliceClient.GetIdentityConfig(aliceClient.Ctx(), &identity.GetIdentityConfigRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call GetIdentityConfig", alice), err.Error())

	_, err = aliceClient.CreateIDPConnector(aliceClient.Ctx(), &identity.CreateIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call CreateIDPConnector", alice), err.Error())

	_, err = aliceClient.GetIDPConnector(aliceClient.Ctx(), &identity.GetIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call GetIDPConnector", alice), err.Error())

	_, err = aliceClient.UpdateIDPConnector(aliceClient.Ctx(), &identity.UpdateIDPConnectorRequest{})
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call UpdateIDPConnector", alice), err.Error())
	require.YesError(t, err)

	_, err = aliceClient.ListIDPConnectors(aliceClient.Ctx(), &identity.ListIDPConnectorsRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call ListIDPConnectors", alice), err.Error())

	_, err = aliceClient.DeleteIDPConnector(aliceClient.Ctx(), &identity.DeleteIDPConnectorRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call DeleteIDPConnector", alice), err.Error())

	_, err = aliceClient.CreateOIDCClient(aliceClient.Ctx(), &identity.CreateOIDCClientRequest{})
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call CreateOIDCClient", alice), err.Error())
	require.YesError(t, err)

	_, err = aliceClient.GetOIDCClient(aliceClient.Ctx(), &identity.GetOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call GetOIDCClient", alice), err.Error())

	_, err = aliceClient.UpdateOIDCClient(aliceClient.Ctx(), &identity.UpdateOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call UpdateOIDCClient", alice), err.Error())

	_, err = aliceClient.ListOIDCClients(aliceClient.Ctx(), &identity.ListOIDCClientsRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call ListOIDCClients", alice), err.Error())

	_, err = aliceClient.DeleteOIDCClient(aliceClient.Ctx(), &identity.DeleteOIDCClientRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call DeleteOIDCClient", alice), err.Error())

	_, err = aliceClient.IdentityAPIClient.DeleteAll(aliceClient.Ctx(), &identity.DeleteAllRequest{})
	require.YesError(t, err)
	require.Equal(t, fmt.Sprintf("rpc error: code = Unknown desc = github:%v is not authorized to perform this operation; must be an admin to call DeleteAll", alice), err.Error())
}

// TestSetConfiguration tests that the web server configuration reloads when the etcd config value is updated
func TestSetConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	adminClient := tu.GetAuthenticatedPachClient(t, tu.AdminUser)

	// Configure an IDP connector, so the web server will start
	_, err := adminClient.CreateIDPConnector(adminClient.Ctx(), &identity.CreateIDPConnectorRequest{
		Connector: &identity.IDPConnector{
			Id:         "id",
			Name:       "name",
			Type:       "mockPassword",
			JsonConfig: `{"username": "test", "password": "test"}`,
		},
	})
	require.NoError(t, err)

	_, err = adminClient.SetIdentityConfig(adminClient.Ctx(), &identity.SetIdentityConfigRequest{
		Config: &identity.IdentityConfig{
			Issuer: "http://localhost:30658/",
		},
	})
	require.NoError(t, err)

	// Block until the web server has restarted with the right config
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.GetIdentityConfig(adminClient.Ctx(), &identity.GetIdentityConfigRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			"http://localhost:30658/", resp.Config.Issuer,
		)
	}, backoff.NewTestingBackOff()))

	resp, err := http.Get(fmt.Sprintf("http://%v/.well-known/openid-configuration", dexHost(adminClient)))
	require.NoError(t, err)

	var oidcConfig map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&oidcConfig))
	require.Equal(t, "http://localhost:30658/", oidcConfig["issuer"].(string))
}

func TestOIDCClientCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	adminClient := tu.GetAuthenticatedPachClient(t, tu.AdminUser)

	client := &identity.OIDCClient{
		Id:           "id",
		Name:         "name",
		Secret:       "secret",
		RedirectUris: []string{"http://localhost:1234"},
		TrustedPeers: []string{"a", "b", "c"},
	}

	_, err := adminClient.CreateOIDCClient(adminClient.Ctx(), &identity.CreateOIDCClientRequest{
		Client: client,
	})
	require.NoError(t, err)

	client.RedirectUris = []string{"http://localhost:1234/redirect"}
	client.Name = "name2"

	_, err = adminClient.UpdateOIDCClient(adminClient.Ctx(), &identity.UpdateOIDCClientRequest{
		Client: client,
	})
	require.NoError(t, err)

	listResp, err := adminClient.ListOIDCClients(adminClient.Ctx(), &identity.ListOIDCClientsRequest{})
	require.NoError(t, err)
	require.Equal(t, []*identity.OIDCClient{client}, listResp.Clients)

	getResp, err := adminClient.GetOIDCClient(adminClient.Ctx(), &identity.GetOIDCClientRequest{Id: "id"})
	require.NoError(t, err)
	require.Equal(t, client, getResp.Client)

	_, err = adminClient.DeleteOIDCClient(adminClient.Ctx(), &identity.DeleteOIDCClientRequest{Id: "id"})
	require.NoError(t, err)

	listResp, err = adminClient.ListOIDCClients(adminClient.Ctx(), &identity.ListOIDCClientsRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(listResp.Clients))

}

func TestIDPConnectorCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	adminClient := tu.GetAuthenticatedPachClient(t, tu.AdminUser)

	conn := &identity.IDPConnector{
		Id:         "id",
		Name:       "name",
		Type:       "mockPassword",
		JsonConfig: `{"username": "test", "password": "test"}`,
	}

	_, err := adminClient.CreateIDPConnector(adminClient.Ctx(), &identity.CreateIDPConnectorRequest{
		Connector: conn,
	})
	require.NoError(t, err)

	conn.ConfigVersion = 1
	conn.Name = "name2"
	_, err = adminClient.UpdateIDPConnector(adminClient.Ctx(), &identity.UpdateIDPConnectorRequest{
		Connector: conn,
	})
	require.NoError(t, err)

	listResp, err := adminClient.ListIDPConnectors(adminClient.Ctx(), &identity.ListIDPConnectorsRequest{})
	require.NoError(t, err)
	require.Equal(t, []*identity.IDPConnector{conn}, listResp.Connectors)

	getResp, err := adminClient.GetIDPConnector(adminClient.Ctx(), &identity.GetIDPConnectorRequest{Id: "id"})
	require.NoError(t, err)
	require.Equal(t, conn, getResp.Connector)

	_, err = adminClient.DeleteIDPConnector(adminClient.Ctx(), &identity.DeleteIDPConnectorRequest{Id: "id"})
	require.NoError(t, err)

	listResp, err = adminClient.ListIDPConnectors(adminClient.Ctx(), &identity.ListIDPConnectorsRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(listResp.Connectors))
}

func dexHost(c *client.APIClient) string {
	parts := strings.Split(c.GetAddress(), ":")
	if parts[1] == "650" {
		return parts[0] + ":658"
	}
	return parts[0] + ":30658"
}
