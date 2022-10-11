//go:build unit_test

package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// TestAuthNotActivated checks that no RPCs can be made when the auth service is disabled
func TestAuthNotActivated(t *testing.T) {
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	client := env.PachClient
	_, err := client.SetIdentityServerConfig(client.Ctx(), &identity.SetIdentityServerConfigRequest{})
	require.YesError(t, err)
	require.Equal(t, "rpc error: code = Unimplemented desc = the auth service is not activated", err.Error())

	_, err = client.GetIdentityServerConfig(client.Ctx(), &identity.GetIdentityServerConfigRequest{})
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
	alice := tu.UniqueString("robot:alice")
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	aliceClient := tu.AuthenticatedPachClient(t, c, alice, peerPort)
	_, err := aliceClient.SetIdentityServerConfig(aliceClient.Ctx(), &identity.SetIdentityServerConfigRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.GetIdentityServerConfig(aliceClient.Ctx(), &identity.GetIdentityServerConfigRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.CreateIDPConnector(aliceClient.Ctx(), &identity.CreateIDPConnectorRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.GetIDPConnector(aliceClient.Ctx(), &identity.GetIDPConnectorRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.UpdateIDPConnector(aliceClient.Ctx(), &identity.UpdateIDPConnectorRequest{})
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())
	require.YesError(t, err)

	_, err = aliceClient.ListIDPConnectors(aliceClient.Ctx(), &identity.ListIDPConnectorsRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.DeleteIDPConnector(aliceClient.Ctx(), &identity.DeleteIDPConnectorRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.CreateOIDCClient(aliceClient.Ctx(), &identity.CreateOIDCClientRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.GetOIDCClient(aliceClient.Ctx(), &identity.GetOIDCClientRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.UpdateOIDCClient(aliceClient.Ctx(), &identity.UpdateOIDCClientRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.ListOIDCClients(aliceClient.Ctx(), &identity.ListOIDCClientsRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.DeleteOIDCClient(aliceClient.Ctx(), &identity.DeleteOIDCClientRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())

	_, err = aliceClient.IdentityAPIClient.DeleteAll(aliceClient.Ctx(), &identity.DeleteAllRequest{})
	require.YesError(t, err)
	require.Matches(t, fmt.Sprintf("rpc error: code = Unknown desc = %v is not authorized to perform this operation", alice), err.Error())
}

// TestSetConfiguration tests that the web server configuration reloads when the etcd config value is updated
func TestSetConfiguration(t *testing.T) {
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	adminClient := tu.AuthenticatedPachClient(t, c, auth.RootUser, peerPort)

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

	_, err = adminClient.SetIdentityServerConfig(adminClient.Ctx(), &identity.SetIdentityServerConfigRequest{
		Config: &identity.IdentityServerConfig{
			Issuer: "http://localhost:30658/dex",
		},
	})
	require.NoError(t, err)

	// Block until the web server has restarted with the right config
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.GetIdentityServerConfig(adminClient.Ctx(), &identity.GetIdentityServerConfigRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			"http://localhost:30658/dex", resp.Config.Issuer,
		)
	}, backoff.NewTestingBackOff()))

	resp, err := http.Get(fmt.Sprintf("http://%v/dex/.well-known/openid-configuration", tu.DexHost(adminClient)))
	require.NoError(t, err)

	var oidcConfig map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&oidcConfig))
	require.Equal(t, "http://localhost:30658/dex", oidcConfig["issuer"].(string))
}

func TestOIDCClientCRUD(t *testing.T) {
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	adminClient := tu.AuthenticatedPachClient(t, c, auth.RootUser, peerPort)

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
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	adminClient := tu.AuthenticatedPachClient(t, c, auth.RootUser, peerPort)

	conn := &identity.IDPConnector{
		Id:   "id",
		Name: "name",
		Type: "mockPassword",
		Config: &types.Struct{
			Fields: map[string]*types.Value{
				"password": {Kind: &types.Value_StringValue{StringValue: "test"}},
				"username": {Kind: &types.Value_StringValue{StringValue: "test"}},
			},
		},
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

// TestShortenIDTokenExpiry tests that we can configure Dex to issue ID tokens with a
// expiration shorter than the default of 6 hours
func TestShortenIDTokenExpiry(t *testing.T) {
	env := realenv.NewRealEnvWithIdentity(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	tu.ActivateAuthClient(t, c, peerPort)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c, true))
	adminClient := tu.AuthenticateClient(t, c, auth.RootUser)
	issuerHost := c.GetAddress().Host
	issuerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort + 8))
	_, err := adminClient.SetIdentityServerConfig(adminClient.Ctx(), &identity.SetIdentityServerConfigRequest{
		Config: &identity.IdentityServerConfig{
			Issuer:              "http://" + issuerHost + ":" + issuerPort + "/dex",
			IdTokenExpiry:       "1h",
			RotationTokenExpiry: "5h",
		},
	})
	require.NoError(t, err)

	token := tu.GetOIDCTokenForTrustedApp(t, c, true)

	// Exchange the ID token for a pach token and confirm the expiration is < 1h
	testClient := tu.UnauthenticatedPachClient(t, c)
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{IdToken: token})
	require.NoError(t, err)

	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.True(t, time.Until(*whoAmIResp.Expiration) < time.Hour)
}
