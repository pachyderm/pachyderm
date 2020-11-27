package server

import (
	"context"
	"net/http"
	"path"
	"sync"
	"time"

	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
)

const (
	dexHTTPPort = ":658"

	configPrefix = "/config"
	configKey    = "config"
)

type apiServer struct {
	pachLogger log.Logger
	env        *serviceenv.ServiceEnv

	config         col.Collection
	configCacheMtx sync.RWMutex
	configCache    identity.IdentityConfig

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
		config: col.NewCollection(
			env.GetEtcdClient(),
			path.Join(etcdPrefix, configPrefix),
			nil,
			&identity.IdentityConfig{},
			nil,
			nil,
		),
	}

	if public {
		server.web = newDexWeb(sp, logger)
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
		var config identity.IdentityConfig
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

func (a *apiServer) SetIdentityConfig(ctx context.Context, req *identity.SetIdentityConfigRequest) (resp *identity.SetIdentityConfigResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "SetIdentityConfig"); err != nil {
		return nil, err
	}

	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		return a.config.ReadWrite(stm).Put(configKey, req.Config)
	}); err != nil {
		return nil, err
	}
	return &identity.SetIdentityConfigResponse{}, nil
}

func (a *apiServer) GetIdentityConfig(ctx context.Context, req *identity.GetIdentityConfigRequest) (resp *identity.GetIdentityConfigResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "GetIdentityConfig"); err != nil {
		return nil, err
	}

	// Serve the cached version from the watcher - this ensures the version matches the one being used by the web server
	a.configCacheMtx.RLock()
	defer a.configCacheMtx.RUnlock()

	return &identity.GetIdentityConfigResponse{Config: &a.configCache}, nil
}

func (a *apiServer) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (resp *identity.CreateIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "CreateIDPConnector"); err != nil {
		return nil, err
	}

	if err := a.api.createConnector(req); err != nil {
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

	c, err := a.api.getConnector(req.Id)
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

	if err := a.api.updateConnector(req); err != nil {
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

	connectors, err := a.api.listConnectors()
	if err != nil {
		return nil, err
	}

	return &identity.ListIDPConnectorsResponse{Connectors: connectors}, nil
}

func (a *apiServer) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (resp *identity.DeleteIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteIDPConnector"); err != nil {
		return nil, err
	}

	if err := a.api.deleteConnector(req.Id); err != nil {
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

	if err := a.isAdmin(ctx, "UpdateOIDCClient"); err != nil {
		return nil, err
	}

	if err := a.api.updateClient(ctx, req); err != nil {
		return nil, err
	}

	return &identity.UpdateOIDCClientResponse{}, nil
}

func (a *apiServer) GetOIDCClient(ctx context.Context, req *identity.GetOIDCClientRequest) (resp *identity.GetOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "GetOIDCClient"); err != nil {
		return nil, err
	}

	client, err := a.api.getClient(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetOIDCClientResponse{Client: client}, nil
}

func (a *apiServer) ListOIDCClients(ctx context.Context, req *identity.ListOIDCClientsRequest) (resp *identity.ListOIDCClientsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "ListOIDCClient"); err != nil {
		return nil, err
	}

	clients, err := a.api.listClients()
	if err != nil {
		return nil, err
	}

	resp = &identity.ListOIDCClientsResponse{Clients: make([]*identity.OIDCClient, len(clients))}

	return resp, nil
}

func (a *apiServer) DeleteOIDCClient(ctx context.Context, req *identity.DeleteOIDCClientRequest) (resp *identity.DeleteOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteOIDCClient"); err != nil {
		return nil, err
	}

	if err := a.api.deleteClient(ctx, req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteOIDCClientResponse{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *identity.DeleteAllRequest) (resp *identity.DeleteAllResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isAdmin(ctx, "DeleteAll"); err != nil {
		return nil, err
	}

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
