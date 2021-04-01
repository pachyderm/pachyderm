package server

import (
	"context"
	"net/http"
	"path"
	"sync"
	"time"

	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

const (
	dexHTTPPort = ":658"

	configPrefix = "/config"
	configKey    = "config"
)

type apiServer struct {
	pachLogger log.Logger
	env        *serviceenv.ServiceEnv

	config         col.EtcdCollection
	configCacheMtx sync.RWMutex
	configCache    identity.IdentityServerConfig

	api *dexAPI
	web *dexWeb
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

// NewIdentityServer returns an implementation of identity.APIServer.
func NewIdentityServer(env *serviceenv.ServiceEnv, sp StorageProvider, public bool, etcdPrefix string) (identity.APIServer, error) {
	logger := logrus.NewEntry(logrus.New()).WithField("source", "dex")

	server := &apiServer{
		env:        env,
		pachLogger: log.NewLogger("identity.API"),
		api:        newDexAPI(sp, logger),
		config: col.NewEtcdCollection(
			env.GetEtcdClient(),
			path.Join(etcdPrefix, configPrefix),
			nil,
			&identity.IdentityServerConfig{},
			nil,
			nil,
		),
	}

	if public {
		server.web = newDexWeb(sp, logger, env.GetDBClient())
		go func() {
			if err := http.ListenAndServe(dexHTTPPort, server.web); err != nil {
				logger.WithError(err).Error("Dex web server stopped")
			}
		}()

		go server.watchConfig()
	}

	return server, nil
}

func (a *apiServer) watchConfig() {
	b := backoff.NewExponentialBackOff()
	backoff.RetryNotify(func() error {
		configWatcher, err := a.config.ReadOnly(context.Background()).Watch()
		if err != nil {
			return err
		}
		defer configWatcher.Close()

		var key string
		var config identity.IdentityServerConfig
		for {
			ev, ok := <-configWatcher.Watch()
			if !ok {
				return errors.New("admin watch closed unexpectedly")
			}
			b.Reset()

			ev.Unmarshal(&key, &config)

			a.configCacheMtx.Lock()
			a.web.updateConfig(config)
			a.configCache = config
			a.configCacheMtx.Unlock()
		}
	}, b, func(err error, d time.Duration) error {
		logrus.Errorf("error watching identity config: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *apiServer) SetIdentityServerConfig(ctx context.Context, req *identity.SetIdentityServerConfigRequest) (resp *identity.SetIdentityServerConfigResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		return a.config.ReadWrite(stm).Put(configKey, req.Config)
	}); err != nil {
		return nil, err
	}
	return &identity.SetIdentityServerConfigResponse{}, nil
}

func (a *apiServer) GetIdentityServerConfig(ctx context.Context, req *identity.GetIdentityServerConfigRequest) (resp *identity.GetIdentityServerConfigResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	// Serve the cached version from the watcher - this ensures the version matches the one being used by the web server
	a.configCacheMtx.RLock()
	defer a.configCacheMtx.RUnlock()

	return &identity.GetIdentityServerConfigResponse{Config: &a.configCache}, nil
}

func (a *apiServer) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (resp *identity.CreateIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.createConnector(req); err != nil {
		return nil, err
	}
	return &identity.CreateIDPConnectorResponse{}, nil
}

func (a *apiServer) GetIDPConnector(ctx context.Context, req *identity.GetIDPConnectorRequest) (resp *identity.GetIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	c, err := a.api.getConnector(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetIDPConnectorResponse{
		Connector: c,
	}, nil
}

func (a *apiServer) UpdateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (resp *identity.UpdateIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.updateConnector(req); err != nil {
		return nil, err
	}

	return &identity.UpdateIDPConnectorResponse{}, nil
}

func (a *apiServer) ListIDPConnectors(ctx context.Context, req *identity.ListIDPConnectorsRequest) (resp *identity.ListIDPConnectorsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	connectors, err := a.api.listConnectors()
	if err != nil {
		return nil, err
	}

	return &identity.ListIDPConnectorsResponse{Connectors: connectors}, nil
}

func (a *apiServer) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (resp *identity.DeleteIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.deleteConnector(req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteIDPConnectorResponse{}, nil
}

func (a *apiServer) CreateOIDCClient(ctx context.Context, req *identity.CreateOIDCClientRequest) (resp *identity.CreateOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	client, err := a.api.createClient(ctx, req)
	if err != nil {
		return nil, err
	}

	return &identity.CreateOIDCClientResponse{
		Client: client,
	}, nil
}

func (a *apiServer) UpdateOIDCClient(ctx context.Context, req *identity.UpdateOIDCClientRequest) (resp *identity.UpdateOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.updateClient(ctx, req); err != nil {
		return nil, err
	}

	return &identity.UpdateOIDCClientResponse{}, nil
}

func (a *apiServer) GetOIDCClient(ctx context.Context, req *identity.GetOIDCClientRequest) (resp *identity.GetOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	client, err := a.api.getClient(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetOIDCClientResponse{Client: client}, nil
}

func (a *apiServer) ListOIDCClients(ctx context.Context, req *identity.ListOIDCClientsRequest) (resp *identity.ListOIDCClientsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	clients, err := a.api.listClients()
	if err != nil {
		return nil, err
	}

	return &identity.ListOIDCClientsResponse{Clients: clients}, nil
}

func (a *apiServer) DeleteOIDCClient(ctx context.Context, req *identity.DeleteOIDCClientRequest) (resp *identity.DeleteOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.deleteClient(ctx, req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteOIDCClientResponse{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *identity.DeleteAllRequest) (resp *identity.DeleteAllResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	clients, err := a.api.listClients()
	if err != nil {
		return nil, err
	}

	connectors, err := a.api.listConnectors()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if err := a.api.deleteClient(ctx, client.Id); err != nil {
			return nil, err
		}
	}

	for _, conn := range connectors {
		if err := a.api.deleteConnector(conn.Id); err != nil {
			return nil, err
		}
	}

	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		if err := a.config.ReadWrite(stm).Delete(configKey); err != nil && !col.IsErrNotFound(err) {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &identity.DeleteAllResponse{}, nil
}
