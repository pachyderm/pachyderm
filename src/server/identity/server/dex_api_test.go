package server

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/server/identityutil"

	dex_memory "github.com/dexidp/dex/storage/memory"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestConnectorCreateListGet tests creating, listing, getting and deleting IDP connectors.
func TestConnectorCreateListGetDelete(t *testing.T) {
	conn1 := &identity.IDPConnector{
		Id:            "conn1",
		Name:          "name1",
		Type:          "github",
		ConfigVersion: 0,
		JsonConfig:    "{}",
	}

	conn2 := &identity.IDPConnector{
		Id:            "conn2",
		Name:          "name2",
		Type:          "github",
		ConfigVersion: 0,
		Config: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	// Create a new connector
	err := api.createConnector(&identity.CreateIDPConnectorRequest{Connector: conn1})
	require.NoError(t, err)

	// Attempt to reuse the connector ID
	err = api.createConnector(&identity.CreateIDPConnectorRequest{Connector: conn1})
	require.YesError(t, err)

	// Create a second connector with a unique ID
	err = api.createConnector(&identity.CreateIDPConnectorRequest{Connector: conn2})
	require.NoError(t, err)

	// List connectors
	connectors, err := api.listConnectors()
	require.NoError(t, err)
	require.Equal(t, 2, len(connectors))

	// Get the properties of a single connector
	getConn, err := api.getConnector(conn2.Id)
	require.NoError(t, err)
	require.Equal(t, conn2, getConn)

	// Delete a connector
	err = api.deleteConnector(conn1.Id)
	require.NoError(t, err)

	// Confirm the connector is deleted
	connectors, err = api.listConnectors()
	require.NoError(t, err)
	require.Equal(t, []*identity.IDPConnector{conn2}, connectors)
}

func mustStruct(in map[string]any) *structpb.Struct {
	result, err := structpb.NewStruct(in)
	if err != nil {
		panic(err)
	}
	return result
}

// TestCreateConnector tests that connects with valid configs can be created.
func TestCreateConnector(t *testing.T) {
	cases := []*identity.IDPConnector{
		{
			Id:            "id1",
			Name:          "name",
			Type:          "github",
			ConfigVersion: 0,
			JsonConfig:    `{"username": "username", "password": "password"}`,
		},
		{
			Id:            "id1",
			Name:          "name",
			Type:          "github",
			ConfigVersion: 0,
			Config:        &structpb.Struct{}, // equivalent to an empty config, in which case jsonConfig should be used.
			JsonConfig:    `{"username": "username", "password": "password"}`,
		},
		{
			Id:            "id1",
			Name:          "name",
			ConfigVersion: 0,
			Type:          "github",
			Config: mustStruct(
				map[string]any{
					"password": "password",
					"username": "username",
				},
			),
		},
		{
			Id:            "id1",
			Name:          "name",
			ConfigVersion: 0,
			Type:          "github",
			Config: mustStruct(
				map[string]any{
					"password": "password",
					"username": "username",
				},
			),
			JsonConfig: "{ invalid values in jsonConfig should be ignored if a valid json config is defined.",
		},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	for _, c := range cases {
		t.Run("valid connectors", func(t *testing.T) {
			err := api.createConnector(&identity.CreateIDPConnectorRequest{Connector: c})
			require.NoError(t, err)
			err = api.deleteConnector(c.Id)
			require.NoError(t, err)
		})
	}
}

// TestCreateInvalidConnector tests that connectors with bad config cannot be created
func TestCreateInvalidConnector(t *testing.T) {
	cases := []struct {
		conn *identity.IDPConnector
		err  string
	}{
		{
			conn: &identity.IDPConnector{
				Name:          "name",
				Type:          "github",
				ConfigVersion: 0,
				JsonConfig:    "{}",
			},
			err: "no id specified",
		},
		{
			conn: &identity.IDPConnector{
				Id:            "id1",
				Type:          "github",
				ConfigVersion: 0,
				JsonConfig:    "{}",
			},
			err: "no name specified",
		},
		{
			conn: &identity.IDPConnector{
				Id:            "id1",
				Name:          "name",
				ConfigVersion: 0,
				JsonConfig:    "{}",
			},
			err: "no type specified",
		},
		{
			conn: &identity.IDPConnector{
				Id:            "id1",
				Name:          "name",
				ConfigVersion: 0,
				Type:          "weirdType",
				JsonConfig:    "{}",
			},
			err: `unknown connector type "weirdType"`,
		},
		{
			conn: &identity.IDPConnector{
				Id:            "id1",
				Name:          "name",
				ConfigVersion: 0,
				Type:          "github",
				JsonConfig:    "{",
			},
			err: `unable to deserialize JSON: unexpected end of JSON input`,
		},
		{
			conn: &identity.IDPConnector{
				Id:            "id1",
				Name:          "name",
				ConfigVersion: 0,
				Type:          "github",
				Config:        &structpb.Struct{},
			},
			err: identityutil.NoConfigErr,
		},
		{
			conn: &identity.IDPConnector{
				Id:            "id1",
				Name:          "name",
				ConfigVersion: 0,
				Type:          "github",
				Config:        &structpb.Struct{}, // this 'empty' value should be overridden by jsonConfig.
				JsonConfig:    "{",
			},
			err: `unable to deserialize JSON: unexpected end of JSON input`,
		},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	for _, c := range cases {
		t.Run("invalid connectors", func(t *testing.T) {
			err := api.createConnector(&identity.CreateIDPConnectorRequest{Connector: c.conn})
			require.YesError(t, err)
			require.Equal(t, c.err, err.Error())
		})
	}
}

// TestUpdateConnector tests updating the individual connector fields.
func TestUpdateConnector(t *testing.T) {
	conn := &identity.IDPConnector{
		Id:            "conn",
		Name:          "name",
		Type:          "github",
		ConfigVersion: 0,
		JsonConfig:    "{}",
	}

	cases := []struct {
		req      *identity.IDPConnector
		expected *identity.IDPConnector
		err      string
	}{
		{
			req: &identity.IDPConnector{
				Id:            "conn",
				Name:          "newname",
				ConfigVersion: 1,
			},
			expected: &identity.IDPConnector{
				Id:            "conn",
				Name:          "newname",
				Type:          "github",
				ConfigVersion: 1,
				Config: &structpb.Struct{
					Fields: map[string]*structpb.Value{},
				},
			},
		},
		{
			req: &identity.IDPConnector{
				Name:          "newername",
				ConfigVersion: 1,
			},
			err: "not found",
		},
		{
			req: &identity.IDPConnector{
				Id:            "conn",
				Name:          "newername",
				ConfigVersion: 1,
			},
			err: "new config version is 1, expected 2",
		},
		{
			req: &identity.IDPConnector{
				Id:            "conn",
				JsonConfig:    "{",
				ConfigVersion: 2,
			},
			err: "unable to deserialize JSON: unexpected end of JSON input",
		},
		{
			req: &identity.IDPConnector{
				Id:            "conn",
				JsonConfig:    `{"client_id": "1234"}`,
				ConfigVersion: 2,
			},
			expected: &identity.IDPConnector{
				Id:            "conn",
				Name:          "newname",
				Type:          "github",
				ConfigVersion: 2,
				Config: mustStruct(
					map[string]any{
						"client_id": "1234",
					},
				),
			},
		},
		{
			req: &identity.IDPConnector{
				Id:            "conn",
				Type:          "mockPassword",
				ConfigVersion: 3,
			},
			err: "unable to open connector: no username supplied",
		},
		{
			req: &identity.IDPConnector{
				Id:            "conn",
				Type:          "mockPassword",
				JsonConfig:    `{"username": "user", "password": "pass"}`,
				ConfigVersion: 3,
			},
			expected: &identity.IDPConnector{
				Id:            "conn",
				Name:          "newname",
				Type:          "mockPassword",
				ConfigVersion: 3,
				Config: mustStruct(
					map[string]any{
						"password": "pass",
						"username": "user",
					},
				),
			},
		},
		{
			req: &identity.IDPConnector{
				Id:   "conn",
				Type: "mockPassword",
				Config: mustStruct(
					map[string]any{
						"password": "pass",
						"username": "user",
					},
				),
				ConfigVersion: 4,
			},
			expected: &identity.IDPConnector{
				Id:            "conn",
				Name:          "newname",
				Type:          "mockPassword",
				ConfigVersion: 4,
				Config: mustStruct(
					map[string]any{
						"password": "pass",
						"username": "user",
					},
				),
			},
		},
		{
			req: &identity.IDPConnector{
				Id:   "conn",
				Type: "mockPassword",
				Config: mustStruct(
					map[string]any{
						"garbageIn": "garbageOut",
					},
				),
				ConfigVersion: 5,
			},
			err: "unable to open connector: no username supplied",
		},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	// Create the initial  connector
	err := api.createConnector(&identity.CreateIDPConnectorRequest{Connector: conn})
	require.NoError(t, err)

	for _, c := range cases {
		t.Run("update connector", func(t *testing.T) {
			err := api.updateConnector(&identity.UpdateIDPConnectorRequest{Connector: c.req})
			if c.err != "" {
				require.YesError(t, err)
				require.Equal(t, c.err, err.Error())
				return
			}
			require.NoError(t, err)

			actual, err := api.getConnector(conn.Id)
			require.NoError(t, err)
			require.Equal(t, c.expected, actual)
		})
	}
}

// TestClientCreateListGet tests creating, listing, getting and deleting OIDC clients.
func TestClientCreateListGetDelete(t *testing.T) {
	client1 := &identity.OIDCClient{
		Id:           "client1",
		Name:         "name1",
		Secret:       "secret1",
		RedirectUris: []string{"http://example.com/1"},
		TrustedPeers: []string{"client2"},
	}

	client2 := &identity.OIDCClient{
		Id:           "client2",
		Name:         "name2",
		RedirectUris: []string{"http://example.com/2"},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	// Create a new connector
	resp, err := api.createClient(context.Background(), &identity.CreateOIDCClientRequest{Client: client1})
	require.NoError(t, err)
	require.Equal(t, client1, resp)

	// Attempt to reuse the client ID
	_, err = api.createClient(context.Background(), &identity.CreateOIDCClientRequest{Client: client1})
	require.YesError(t, err)

	// Create a second client with a unique ID
	resp, err = api.createClient(context.Background(), &identity.CreateOIDCClientRequest{Client: client2})
	require.NoError(t, err)
	client2.Secret = resp.Secret
	require.Equal(t, client2, resp)

	// List clients
	clients, err := api.listClients()
	require.NoError(t, err)
	require.Equal(t, 2, len(clients))

	// Get the properties of a single client
	getClient, err := api.getClient(client2.Id)
	require.NoError(t, err)
	require.Equal(t, client2, getClient)

	// Delete a client
	err = api.deleteClient(context.Background(), client1.Id)
	require.NoError(t, err)

	// Confirm the client is deleted
	clients, err = api.listClients()
	require.NoError(t, err)
	require.Equal(t, []*identity.OIDCClient{client2}, clients)
}

// TestCreateInvalidClient tests that clients with bad config cannot be created
func TestCreateInvalidClient(t *testing.T) {
	cases := []struct {
		client *identity.OIDCClient
		err    string
	}{
		{
			client: &identity.OIDCClient{
				Id: "client1",
			},
			err: "no client name specified",
		},
		{
			client: &identity.OIDCClient{
				Name: "client1",
			},
			err: "no client id specified",
		},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	for _, c := range cases {
		t.Run("invalid client", func(t *testing.T) {
			_, err := api.createClient(context.Background(), &identity.CreateOIDCClientRequest{Client: c.client})
			require.YesError(t, err)
			require.Equal(t, c.err, err.Error())
		})
	}
}

// TestUpdateClient tests updating the individual client fields
func TestUpdateClient(t *testing.T) {
	client := &identity.OIDCClient{
		Id:           "client",
		Name:         "name",
		Secret:       "secret",
		RedirectUris: []string{"http://example.com/1"},
		TrustedPeers: []string{"a"},
	}

	cases := []struct {
		req      *identity.OIDCClient
		expected *identity.OIDCClient
		err      string
	}{
		{
			req: &identity.OIDCClient{
				Id:   "unknown",
				Name: "newname",
			},
			err: `unable to find OIDC client with id "unknown"`,
		},
		{
			req: &identity.OIDCClient{
				Id:   "client",
				Name: "newname",
			},
			expected: &identity.OIDCClient{
				Id:           "client",
				Name:         "newname",
				Secret:       "secret",
				RedirectUris: []string{"http://example.com/1"},
				TrustedPeers: []string{"a"},
			},
		},
		{
			req: &identity.OIDCClient{
				Id:           "client",
				RedirectUris: []string{"http://example.com/2", "http://example.com/3"},
			},
			expected: &identity.OIDCClient{
				Id:           "client",
				Name:         "newname",
				Secret:       "secret",
				RedirectUris: []string{"http://example.com/2", "http://example.com/3"},
				TrustedPeers: []string{"a"},
			},
		},
		{
			req: &identity.OIDCClient{
				Id:           "client",
				TrustedPeers: []string{"b", "c"},
			},
			expected: &identity.OIDCClient{
				Id:           "client",
				Name:         "newname",
				Secret:       "secret",
				RedirectUris: []string{"http://example.com/2", "http://example.com/3"},
				TrustedPeers: []string{"b", "c"},
			},
		},
		{
			req: &identity.OIDCClient{
				Id:           "client",
				RedirectUris: []string{},
				TrustedPeers: []string{},
			},
			expected: &identity.OIDCClient{
				Id:           "client",
				Name:         "newname",
				Secret:       "secret",
				RedirectUris: []string{},
				TrustedPeers: []string{},
			},
		},
	}

	ctx := pctx.TestContext(t)
	logger := log.NewLogrus(ctx)
	api := newDexAPI(dex_memory.New(logger))

	// Create the initial  connector
	_, err := api.createClient(context.Background(), &identity.CreateOIDCClientRequest{Client: client})
	require.NoError(t, err)

	for _, c := range cases {
		t.Run("update client", func(t *testing.T) {
			err := api.updateClient(context.Background(), &identity.UpdateOIDCClientRequest{Client: c.req})
			if c.err != "" {
				require.YesError(t, err)
				require.Equal(t, c.err, err.Error())
				return
			}
			require.NoError(t, err)

			actual, err := api.getClient(client.Id)
			require.NoError(t, err)
			require.Equal(t, c.expected, actual)
		})
	}
}
