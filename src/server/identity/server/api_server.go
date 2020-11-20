package server

import (
	"context"
	"strconv"
	"time"

	dex_api "github.com/dexidp/dex/api/v2"
	"github.com/google/uuid"
	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
)

type apiServer struct {
	pachLogger log.Logger
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

func NewIdentityServer(pgHost, pgDatabase, pgUser, pgPwd, pgSSL, issuer string, pgPort int, public bool) (*apiServer, error) {
	server, err := newDexServer(pgHost, pgDatabase, pgUser, pgPwd, pgSSL, issuer, pgPort, public)
	if err != nil {
		return nil, err
	}

	return &apiServer{
		pachLogger: log.NewLogger("identity.API"),
		server:     server,
	}, nil
}

func (a *apiServer) CreateConnector(ctx context.Context, req *identity.CreateConnectorRequest) (resp *identity.CreateConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.server.addConnector(req.Config.Id, req.Config.Name, req.Config.Type, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}
	return &identity.CreateConnectorResponse{}, nil
}

func (a *apiServer) UpdateConnector(ctx context.Context, req *identity.UpdateConnectorRequest) (resp *identity.UpdateConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.server.addConnector(req.Config.Id, req.Config.Name, req.Config.Type, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}
	return &identity.UpdateConnectorResponse{}, nil
}

func (a *apiServer) ListConnectors(ctx context.Context, req *identity.ListConnectorsRequest) (resp *identity.ListConnectorsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	connectors, err := a.server.listConnectors()
	if err != nil {
		return nil, err
	}

	resp = &identity.ListConnectorsResponse{
		Config: make([]*identity.ConnectorConfig, len(connectors)),
	}

	for i, c := range connectors {
		version, _ := strconv.Atoi(c.ResourceVersion)
		resp.Config[i] = &identity.ConnectorConfig{
			Id:            c.ID,
			Name:          c.Name,
			Type:          c.Type,
			ConfigVersion: int64(version),
		}
	}

	return resp, nil
}

func (a *apiServer) DeleteConnector(ctx context.Context, req *identity.DeleteConnectorRequest) (resp *identity.DeleteConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.server.deleteConnector(req.Id); err != nil {
		return nil, err
	}
	return &identity.DeleteConnectorResponse{}, nil
}

func (a *apiServer) CreateClient(ctx context.Context, req *identity.CreateClientRequest) (resp *identity.CreateClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	secret := uuid.New().String()

	client := &dex_api.CreateClientReq{
		Client: &dex_api.Client{
			Id:           req.Client.Id,
			Secret:       secret,
			RedirectUris: req.Client.RedirectUris,
			TrustedPeers: req.Client.TrustedPeers,
			Name:         req.Client.Name,
		},
	}

	if _, err := a.server.CreateClient(ctx, client); err != nil {
		return nil, err
	}

	return &identity.CreateClientResponse{
		Secret: secret,
	}, nil
}

func (a *apiServer) DeleteClient(ctx context.Context, req *identity.DeleteClientRequest) (resp *identity.DeleteClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if _, err := a.server.DeleteClient(ctx, &dex_api.DeleteClientReq{Id: req.Id}); err != nil {
		return nil, err
	}
	return &identity.DeleteClientResponse{}, nil
}
