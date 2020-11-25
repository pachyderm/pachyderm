package server

import (
	"context"
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

func dexConnectorToPach(c storage.Connector) *identity.ConnectorConfig {
	// If the version isn't an int, set it to zero
	version, _ := strconv.Atoi(c.ResourceVersion)
	return &identity.ConnectorConfig{
		Id:            c.ID,
		Name:          c.Name,
		Type:          c.Type,
		ConfigVersion: int64(version),
		JsonConfig:    string(c.Config),
	}
}

func dexClientToPach(c *dex_api.Client) *identity.Client {
	return &identity.Client{
		Id:           c.Id,
		Secret:       c.Secret,
		RedirectUris: c.RedirectUris,
		TrustedPeers: c.TrustedPeers,
		Name:         c.Name,
	}
}

func (a *apiServer) CreateConnector(ctx context.Context, req *identity.CreateConnectorRequest) (resp *identity.CreateConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "CreateConnector"); err != nil {
		return nil, err
	}

	if err := a.server.createConnector(req.Config.Id, req.Config.Name, req.Config.Type, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}
	return &identity.CreateConnectorResponse{}, nil
}

func (a *apiServer) GetConnector(ctx context.Context, req *identity.GetConnectorRequest) (resp *identity.GetConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "GetConnector"); err != nil {
		return nil, err
	}

	c, err := a.server.getConnector(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetConnectorResponse{
		Config: dexConnectorToPach(c),
	}, nil
}

func (a *apiServer) UpdateConnector(ctx context.Context, req *identity.UpdateConnectorRequest) (resp *identity.UpdateConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "UpdateConnector"); err != nil {
		return nil, err
	}

	if err := a.server.updateConnector(req.Config.Id, req.Config.Name, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}
	return &identity.UpdateConnectorResponse{}, nil
}

func (a *apiServer) ListConnectors(ctx context.Context, req *identity.ListConnectorsRequest) (resp *identity.ListConnectorsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "ListConnectors"); err != nil {
		return nil, err
	}

	connectors, err := a.server.listConnectors()
	if err != nil {
		return nil, err
	}

	resp = &identity.ListConnectorsResponse{
		Connectors: make([]*identity.ConnectorConfig, len(connectors)),
	}

	for i, c := range connectors {
		resp.Connectors[i] = dexConnectorToPach(c)
	}

	return resp, nil
}

func (a *apiServer) DeleteConnector(ctx context.Context, req *identity.DeleteConnectorRequest) (resp *identity.DeleteConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteConnector"); err != nil {
		return nil, err
	}

	if err := a.server.deleteConnector(req.Id); err != nil {
		return nil, err
	}
	return &identity.DeleteConnectorResponse{}, nil
}

func (a *apiServer) CreateClient(ctx context.Context, req *identity.CreateClientRequest) (resp *identity.CreateClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "CreateClient"); err != nil {
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

	return &identity.CreateClientResponse{
		Client: dexClientToPach(dexResp.Client),
	}, nil
}

func (a *apiServer) DeleteClient(ctx context.Context, req *identity.DeleteClientRequest) (resp *identity.DeleteClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteClient"); err != nil {
		return nil, err
	}

	if _, err := a.server.DeleteClient(ctx, &dex_api.DeleteClientReq{Id: req.Id}); err != nil {
		return nil, err
	}
	return &identity.DeleteClientResponse{}, nil
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
		if _, err := a.server.DeleteClient(ctx, &dex_api.DeleteClientReq{Id: client.ID}); err != nil {
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
