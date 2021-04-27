package server

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/keycache"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
)

const (
	// enterpriseTokenKey is the constant key we use that maps to an Enterprise
	// token that a user has given us. This is what we check to know if a
	// Pachyderm cluster supports enterprise features
	enterpriseTokenKey = "token"
	configKey          = "config"

	heartbeatFrequency = time.Hour
	heartbeatTimeout   = time.Minute
)

type apiServer struct {
	pachLogger log.Logger
	env        serviceenv.ServiceEnv

	enterpriseTokenCache *keycache.Cache

	// enterpriseTokenCol is a collection containing the enterprise license state
	enterpriseTokenCol col.Collection

	// configCol is a collection containing the license server configuration
	configCol col.Collection
}

func (a *apiServer) LogReq(request interface{}) {
	a.pachLogger.Log(request, nil, nil, 0)
}

// NewEnterpriseServer returns an implementation of ec.APIServer.
func NewEnterpriseServer(env serviceenv.ServiceEnv, etcdPrefix string) (ec.APIServer, error) {
	defaultEnterpriseRecord := &ec.EnterpriseRecord{}
	enterpriseTokenCol := col.NewCollection(
		env.GetEtcdClient(),
		etcdPrefix,
		nil,
		&ec.EnterpriseRecord{},
		nil,
		nil,
	)

	s := &apiServer{
		pachLogger:           log.NewLogger("enterprise.API", env.Logger()),
		env:                  env,
		enterpriseTokenCache: keycache.NewCache(env.Context(), enterpriseTokenCol, enterpriseTokenKey, defaultEnterpriseRecord),
		enterpriseTokenCol:   enterpriseTokenCol,
		configCol:            col.NewCollection(env.GetEtcdClient(), etcdPrefix, nil, &ec.EnterpriseConfig{}, nil, nil),
	}
	go s.enterpriseTokenCache.Watch()
	go s.heartbeatRoutine()
	return s, nil
}

// heartbeatRoutine should  be run in a goroutine and attempts to heartbeat to the license service.
// If the attempt fails and the license server is configured it logs the error.
func (a *apiServer) heartbeatRoutine() {
	for range time.Tick(heartbeatFrequency) {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), heartbeatTimeout)
			defer cancel()
			if err := a.heartbeatIfConfigured(ctx); err != nil {
				logrus.WithError(err).Error("enterprise license heartbeat process failed")
			}
		}()
	}
}

// heartbeatIfConfigured attempts to heartbeat to the currently configured license service.
// If no license service is configured it fails with license.ErrNotActivated.
// If the heartbeat fails with ErrInvalidIDOrSecret it moves the cluster into an inactive state.
func (a *apiServer) heartbeatIfConfigured(ctx context.Context) error {
	// If we can't get the license server address, skip heartbeating
	var config ec.EnterpriseConfig
	if err := a.configCol.ReadOnly(ctx).Get(configKey, &config); err != nil {
		if col.IsErrNotFound(err) {
			return lc.ErrNotActivated
		}
		return err
	}

	// If we can't reach the license server, we assume our license is still valid.
	// However, if we reach the license server and the cluster ID has been deleted then we should
	// deactivate the license on this server.
	resp, err := a.heartbeatToServer(ctx, config.LicenseServer, config.Id, config.Secret)
	if err != nil {
		if lc.IsErrInvalidIDOrSecret(err) {
			logrus.WithError(err).Error("enterprise license heartbeat had invalid id or secret, disabling enterprise")
			_, err = col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
				e := a.enterpriseTokenCol.ReadWrite(stm)
				return e.Put(enterpriseTokenKey, &ec.EnterpriseRecord{
					LastHeartbeat:   types.TimestampNow(),
					HeartbeatFailed: true,
				})
			})
			return err
		}
		return err
	}

	_, err = col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		e := a.enterpriseTokenCol.ReadWrite(stm)
		return e.Put(enterpriseTokenKey, &ec.EnterpriseRecord{
			LastHeartbeat:   types.TimestampNow(),
			License:         resp.License,
			HeartbeatFailed: false,
		})
	})
	return err
}

// heartbeatToServer heartbeats to the provided license server with the id and secret, and updates
// the state in etcd if it's successful.
func (a *apiServer) heartbeatToServer(ctx context.Context, licenseServer, id, secret string) (*lc.HeartbeatResponse, error) {
	localClient := a.env.GetPachClient(ctx)
	versionResp, err := localClient.Version()
	if err != nil {
		return nil, err
	}

	authEnabled, err := localClient.IsAuthActive()
	if err != nil {
		return nil, err
	}

	pachdAddress, err := grpcutil.ParsePachdAddress(licenseServer)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse the active context's pachd address")
	}

	var options []client.Option
	if pachdAddress.Secured {
		options = append(options, client.WithSystemCAs)
	}

	pachClient, err := client.NewFromAddress(pachdAddress.Hostname(), options...)
	if err != nil {
		return nil, err
	}

	return pachClient.License.Heartbeat(ctx, &lc.HeartbeatRequest{
		Id:          id,
		Secret:      secret,
		Version:     versionResp,
		AuthEnabled: authEnabled,
	})
}

// Heartbeat implements the Heartbeat RPC. It exists mostly to test the heartbeat logic
func (a *apiServer) Heartbeat(ctx context.Context, req *ec.HeartbeatRequest) (resp *ec.HeartbeatResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.pachLogger.Log(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.heartbeatIfConfigured(ctx); err != nil {
		return nil, err
	}
	return &ec.HeartbeatResponse{}, nil
}

// Activate implements the Activate RPC
func (a *apiServer) Activate(ctx context.Context, req *ec.ActivateRequest) (resp *ec.ActivateResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.pachLogger.Log(req, resp, retErr, time.Since(start)) }(time.Now())

	// Try to heartbeat before persisting the configuration
	heartbeatResp, err := a.heartbeatToServer(ctx, req.LicenseServer, req.Id, req.Secret)
	if err != nil {
		return nil, err
	}

	record := &ec.EnterpriseRecord{License: heartbeatResp.License}

	// If the test heartbeat succeeded, write the state and config to etcd
	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		if err := a.configCol.ReadWrite(stm).Put(configKey, &ec.EnterpriseConfig{
			LicenseServer: req.LicenseServer,
			Id:            req.Id,
			Secret:        req.Secret,
		}); err != nil {
			return err
		}

		return a.enterpriseTokenCol.ReadWrite(stm).Put(enterpriseTokenKey, record)
	}); err != nil {
		return nil, err
	}

	// Wait until watcher observes the write to the state key
	if err := backoff.Retry(func() error {
		cachedRecord, ok := a.enterpriseTokenCache.Load().(*ec.EnterpriseRecord)
		if !ok {
			return errors.Errorf("could not retrieve enterprise expiration time")
		}
		if !proto.Equal(cachedRecord, record) {
			return errors.Errorf("enterprise not activated")
		}
		return nil
	}, backoff.RetryEvery(100*time.Millisecond)); err != nil {
		return nil, err
	}

	return &ec.ActivateResponse{}, nil
}

// GetState returns the current state of the cluster's Pachyderm Enterprise key (ACTIVE, EXPIRED, or NONE), without the signature of the activation coee
// redacted so it can be used as an identifier but can't be replayed.
func (a *apiServer) GetState(ctx context.Context, req *ec.GetStateRequest) (resp *ec.GetStateResponse, retErr error) {
	record, err := a.getEnterpriseRecord()
	if err != nil {
		return nil, err
	}

	resp = &ec.GetStateResponse{
		Info:  record.Info,
		State: record.State,
	}

	if record.ActivationCode != "" {
		activationCode, err := license.Unmarshal(record.ActivationCode)
		if err != nil {
			return nil, err
		}

		activationCode.Signature = ""
		activationCodeStr, err := json.Marshal(activationCode)
		if err != nil {
			return nil, err
		}

		resp.ActivationCode = base64.StdEncoding.EncodeToString(activationCodeStr)
	}

	return resp, nil
}

// GetActivationCode returns the current state of the cluster's Pachyderm Enterprise key (ACTIVE, EXPIRED, or NONE), including the enterprise activation code
func (a *apiServer) GetActivationCode(ctx context.Context, req *ec.GetActivationCodeRequest) (resp *ec.GetActivationCodeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.pachLogger.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	return a.getEnterpriseRecord()
}

func (a *apiServer) getEnterpriseRecord() (*ec.GetActivationCodeResponse, error) {
	record, ok := a.enterpriseTokenCache.Load().(*ec.EnterpriseRecord)
	if !ok {
		return nil, errors.Errorf("could not retrieve enterprise expiration time")
	}

	if record.License == nil {
		if record.HeartbeatFailed {
			return &ec.GetActivationCodeResponse{
				State: ec.State_HEARTBEAT_FAILED,
			}, nil
		}
		return &ec.GetActivationCodeResponse{
			State: ec.State_NONE,
		}, nil
	}

	expiration, err := types.TimestampFromProto(record.License.Expires)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse expiration timestamp")
	}
	resp := &ec.GetActivationCodeResponse{
		Info: &ec.TokenInfo{
			Expires: record.License.Expires,
		},
		ActivationCode: record.License.ActivationCode,
	}
	if time.Now().After(expiration) {
		resp.State = ec.State_EXPIRED
	} else {
		resp.State = ec.State_ACTIVE
	}
	return resp, nil
}

// Deactivate deletes the current cluster's enterprise token, and puts the
// cluster in the "NONE" enterprise state.
func (a *apiServer) Deactivate(ctx context.Context, req *ec.DeactivateRequest) (resp *ec.DeactivateResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.pachLogger.Log(req, resp, retErr, time.Since(start)) }(time.Now())

	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		err := a.enterpriseTokenCol.ReadWrite(stm).Delete(enterpriseTokenKey)
		if err != nil && !col.IsErrNotFound(err) {
			return err
		}
		err = a.configCol.ReadWrite(stm).Delete(enterpriseTokenKey)
		if err != nil && !col.IsErrNotFound(err) {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Wait until watcher observes the write
	if err := backoff.Retry(func() error {
		record, ok := a.enterpriseTokenCache.Load().(*ec.EnterpriseRecord)
		if !ok {
			return errors.Errorf("could not retrieve enterprise expiration time")
		}
		if record.License != nil {
			return errors.Errorf("enterprise still activated")
		}
		return nil
	}, backoff.RetryEvery(time.Second)); err != nil {
		return nil, err
	}

	return &ec.DeactivateResponse{}, nil
}
