package server

import (
	"context"
	"strconv"
	"time"

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

func NewIdentityServer(etcdAddr, etcdPrefix, issuer string, public bool) (*apiServer, error) {
	server, err := newDexServer(etcdAddr, etcdPrefix, issuer, public)
	if err != nil {
		return nil, err
	}

	return &apiServer{
		pachLogger: log.NewLogger("identity.API"),
		server:     server,
	}, nil
}

func (a *apiServer) AddConnector(ctx context.Context, req *identity.AddConnectorRequest) (resp *identity.AddConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.server.addConnector(req.Config.Id, req.Config.Name, req.Config.Type, int(req.Config.ConfigVersion), []byte(req.Config.JsonConfig)); err != nil {
		return nil, err
	}
	return &identity.AddConnectorResponse{}, nil
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
