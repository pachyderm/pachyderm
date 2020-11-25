package server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	dex_api "github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/storage"
	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

type apiServer struct {
	pachLogger log.Logger
	env        *serviceenv.ServiceEnv
	server     *dexServer
}

// LogReq is like log.Logger.Log(), but it assumes that it's being called from
// the top level of a GRPC method implementation, and correspondingly extracts
// the method name from the parent stack frame
func (a *apiServer) LogReq(request interface{}) {
	a.pachLogger.Log(request, nil, nil, 0)
}

// LogResp is like log.Logger.Log(). However,
// 1) It assumes that it's being called from a defer() statement in a GRPC
//    method , and correspondingly extracts the method name from the grandparent
//    stack frame
func (a *apiServer) LogResp(request interface{}, response interface{}, err error, duration time.Duration) {
	if err == nil {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.InfoLevel, 4)
	} else {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.ErrorLevel, 4)
	}
}

func NewIdentityServer(env *serviceenv.ServiceEnv, pgHost, pgDatabase, pgUser, pgPwd, pgSSL, issuer string, pgPort int, public bool) (*apiServer, error) {
	server, err := newDexServer(pgHost, pgDatabase, pgUser, pgPwd, pgSSL, issuer, pgPort, public)
	if err != nil {
		return nil, err
	}

	return &apiServer{
		env:        env,
		pachLogger: log.NewLogger("identity.API"),
		server:     server,
	}, nil
}

func (a *apiServer) isAdmin(ctx context.Context, op string) error {
	pachClient := a.env.GetPachClient(ctx)
	ctx = pachClient.Ctx() // pachClient will propagate auth info

	// check if the caller is authorized -- they must be an admin
	me, err := pachClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if err != nil {
		return err
	}

	for _, s := range me.ClusterRoles.Roles {
		if s == auth.ClusterRole_SUPER {
			return nil
		}
	}

	return &auth.ErrNotAuthorized{
		Subject: me.Username,
		AdminOp: op,
	}
}

func dexConnectorToPach(c storage.Connector) *identity.IDPConnector {
	// If the version isn't an int, set it to zero
	version, _ := strconv.Atoi(c.ResourceVersion)
	return &identity.IDPConnector{
		Id:            c.ID,
		Name:          c.Name,
		Type:          c.Type,
		ConfigVersion: int64(version),
		JsonConfig:    string(c.Config),
	}
}

func dexClientToPach(c *dex_api.Client) *identity.OIDCClient {
	return &identity.OIDCClient{
		Id:           c.Id,
		Secret:       c.Secret,
		RedirectUris: c.RedirectUris,
		TrustedPeers: c.TrustedPeers,
		Name:         c.Name,
	}
}

func (a *apiServer) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (resp *identity.CreateIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "CreateIDPConnector"); err != nil {
		return nil, err
	}

	if err := a.server.createConnector(req.Config.Id, req.Config.Name, req.Config.Type, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}
	return &identity.CreateIDPConnectorResponse{}, nil
}

func (a *apiServer) GetIDPConnector(ctx context.Context, req *identity.GetIDPConnectorRequest) (resp *identity.GetIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "GetIDPConnector"); err != nil {
		return nil, err
	}

	c, err := a.server.getConnector(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetIDPConnectorResponse{
		Config: dexConnectorToPach(c),
	}, nil
}

func (a *apiServer) UpdateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (resp *identity.UpdateIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "UpdateIDPConnector"); err != nil {
		return nil, err
	}

	if err := a.server.updateConnector(req.Config.Id, req.Config.Name, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}

	return &identity.UpdateIDPConnectorResponse{}, nil
}

func (a *apiServer) ListIDPConnectors(ctx context.Context, req *identity.ListIDPConnectorsRequest) (resp *identity.ListIDPConnectorsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "ListIDPConnectors"); err != nil {
		return nil, err
	}

	connectors, err := a.server.listConnectors()
	if err != nil {
		return nil, err
	}

	resp = &identity.ListIDPConnectorsResponse{
		Connectors: make([]*identity.IDPConnector, len(connectors)),
	}

	for i, c := range connectors {
		resp.Connectors[i] = dexConnectorToPach(c)
	}

	return resp, nil
}

func (a *apiServer) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (resp *identity.DeleteIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteIDPConnector"); err != nil {
		return nil, err
	}

	if err := a.server.deleteConnector(req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteIDPConnectorResponse{}, nil
}

func (a *apiServer) CreateOIDCClient(ctx context.Context, req *identity.CreateOIDCClientRequest) (resp *identity.CreateOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "CreateOIDCClient"); err != nil {
		return nil, err
	}

	client := &dex_api.CreateClientReq{
		Client: &dex_api.Client{
			Id:           req.Client.Id,
			Secret:       req.Client.Secret,
			RedirectUris: req.Client.RedirectUris,
			TrustedPeers: req.Client.TrustedPeers,
			Name:         req.Client.Name,
		},
	}

	dexResp, err := a.server.CreateClient(ctx, client)
	if err != nil {
		return nil, err
	}

	if dexResp.AlreadyExists {
		return nil, fmt.Errorf("OIDC client with id %q already exists", req.Client.Id)
	}

	return &identity.CreateOIDCClientResponse{
		Client: dexClientToPach(dexResp.Client),
	}, nil
}

func (a *apiServer) UpdateOIDCClient(ctx context.Context, req *identity.UpdateOIDCClientRequest) (resp *identity.UpdateOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "UpdateOIDCClient"); err != nil {
		return nil, err
	}

	client := &dex_api.UpdateClientReq{
		Id:           req.Client.Id,
		Name:         req.Client.Name,
		RedirectUris: req.Client.RedirectUris,
		TrustedPeers: req.Client.TrustedPeers,
	}

	dexResp, err := a.server.UpdateClient(ctx, client)
	if err != nil {
		return nil, err
	}

	if dexResp.NotFound {
		return nil, fmt.Errorf("unable to find OIDC client with id %q", req.Client.Id)
	}

	return &identity.UpdateOIDCClientResponse{}, nil
}

func (a *apiServer) GetOIDCClient(ctx context.Context, req *identity.GetOIDCClientRequest) (resp *identity.GetOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "GetOIDCClient"); err != nil {
		return nil, err
	}

	client, err := a.server.getClient(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetOIDCClientResponse{Client: dexClientToPach(client)}, nil
}

func (a *apiServer) ListOIDCClients(ctx context.Context, req *identity.ListOIDCClientsRequest) (resp *identity.ListOIDCClientsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "ListOIDCClient"); err != nil {
		return nil, err
	}

	clients, err := a.server.listClients()
	if err != nil {
		return nil, err
	}

	resp = &identity.ListOIDCClientsResponse{Clients: make([]*identity.OIDCClient, len(clients))}
	for i, client := range clients {
		resp.Clients[i] = dexClientToPach(client)
	}

	return resp, nil
}

func (a *apiServer) DeleteOIDCClient(ctx context.Context, req *identity.DeleteOIDCClientRequest) (resp *identity.DeleteOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteOIDCClient"); err != nil {
		return nil, err
	}

	dexResp, err := a.server.DeleteClient(ctx, &dex_api.DeleteClientReq{Id: req.Id})
	if err != nil {
		return nil, err
	}

	if dexResp.NotFound {
		return nil, fmt.Errorf("unable to find OIDC client with id %q", req.Id)
	}

	return &identity.DeleteOIDCClientResponse{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *identity.DeleteAllRequest) (resp *identity.DeleteAllResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteAll"); err != nil {
		return nil, err
	}

	clients, err := a.server.listClients()
	if err != nil {
		return nil, err
	}

	connectors, err := a.server.listConnectors()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if _, err := a.server.DeleteClient(ctx, &dex_api.DeleteClientReq{Id: client.Id}); err != nil {
			return nil, err
		}
	}

	for _, conn := range connectors {
		if err := a.server.deleteConnector(conn.ID); err != nil {
			return nil, err
		}
	}

	return &identity.DeleteAllResponse{}, nil
}
