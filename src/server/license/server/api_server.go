package server

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/random"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
)

const (
	licenseRecordKey = "license"
)

type apiServer struct {
	pachLogger log.Logger
	env        Env

	// license is the database record where we store the active enterprise license
	license col.PostgresCollection
}

func (a *apiServer) LogReq(ctx context.Context, request interface{}) {
	a.pachLogger.Log(ctx, request, nil, nil, 0)
}

// New returns an implementation of license.APIServer.
func New(env Env) (lc.APIServer, error) {
	s := &apiServer{
		pachLogger: log.NewLogger("license.API", env.Logger),
		env:        env,
		license:    licenseCollection(env.DB, env.Listener),
	}
	return s, nil
}

// Activate implements the Activate RPC
func (a *apiServer) Activate(ctx context.Context, req *lc.ActivateRequest) (resp *lc.ActivateResponse, retErr error) {
	a.LogReq(ctx, nil)
	defer func(start time.Time) { a.pachLogger.Log(ctx, nil, resp, retErr, time.Since(start)) }(time.Now())

	// Validate the activation code
	expiration, err := license.Validate(req.ActivationCode)
	if err != nil {
		return nil, errors.Wrapf(err, "error validating activation code")
	}
	// Allow request to override expiration in the activation code, for testing
	if req.Expires != nil {
		customExpiration, err := types.TimestampFromProto(req.Expires)
		if err == nil && expiration.After(customExpiration) {
			expiration = customExpiration
		}
	}
	expirationProto, err := types.TimestampProto(expiration)
	if err != nil {
		return nil, errors.Wrapf(err, "could not convert expiration time \"%s\" to proto", expiration.String())
	}

	newRecord := &ec.LicenseRecord{
		ActivationCode: req.ActivationCode,
		Expires:        expirationProto,
	}

	if err := dbutil.WithTx(ctx, a.env.DB, func(sqlTx *pachsql.Tx) error {
		return a.license.ReadWrite(sqlTx).Put(licenseRecordKey, newRecord)
	}); err != nil {
		return nil, err
	}

	return &lc.ActivateResponse{
		Info: &ec.TokenInfo{
			Expires: expirationProto,
		},
	}, nil
}

// GetActivationCode returns the current state of the cluster's Pachyderm Enterprise key (ACTIVE, EXPIRED, or NONE), including the enterprise activation code
func (a *apiServer) GetActivationCode(ctx context.Context, req *lc.GetActivationCodeRequest) (resp *lc.GetActivationCodeResponse, retErr error) {
	// Redact the activation code from the response
	removeSecret := func(r *lc.GetActivationCodeResponse) *lc.GetActivationCodeResponse {
		copyResp := proto.Clone(r).(*lc.GetActivationCodeResponse)
		copyResp.ActivationCode = ""
		return copyResp
	}

	a.LogReq(ctx, req)
	defer func(start time.Time) { a.pachLogger.Log(ctx, req, removeSecret(resp), retErr, time.Since(start)) }(time.Now())

	return a.getLicenseRecord(ctx)
}

func (a *apiServer) getLicenseRecord(ctx context.Context) (*lc.GetActivationCodeResponse, error) {
	var record ec.LicenseRecord
	if err := a.license.ReadOnly(ctx).Get(licenseRecordKey, &record); err != nil {
		if col.IsErrNotFound(err) {
			return &lc.GetActivationCodeResponse{State: ec.State_NONE}, nil
		}
		return nil, err
	}

	expiration, err := types.TimestampFromProto(record.Expires)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse expiration timestamp")
	}
	resp := &lc.GetActivationCodeResponse{
		Info: &ec.TokenInfo{
			Expires: record.Expires,
		},
		ActivationCode: record.ActivationCode,
	}
	if time.Now().After(expiration) {
		resp.State = ec.State_EXPIRED
	} else {
		resp.State = ec.State_ACTIVE
	}
	return resp, nil
}

func (a *apiServer) checkLicenseState(ctx context.Context) error {
	record, err := a.getLicenseRecord(ctx)
	if err != nil {
		return err
	}
	if record.State != ec.State_ACTIVE {
		return fmt.Errorf("enterprise license is not valid - %v", record.State)
	}
	return nil
}

func (a *apiServer) validateClusterConfig(ctx context.Context, address string) error {
	if address == "" {
		return errors.New("no address provided for cluster")
	}

	pachClient, err := client.NewFromURI(address)
	if err != nil {
		return errors.Wrapf(err, "unable to create client for %q", address)
	}

	// Make an RPC that doesn't need auth - we don't care about the response, the pachd version is
	// included in the heartbeat
	_, err = pachClient.GetVersion(ctx, &types.Empty{})
	if err != nil {
		return errors.Wrapf(err, "unable to connect to pachd at %q", address)
	}
	return nil
}

// AddCluster registers a new pachd with this license server. Each pachd is configured with a shared secret
// which is used to authenticate to the license server when heartbeating.
func (a *apiServer) AddCluster(ctx context.Context, req *lc.AddClusterRequest) (resp *lc.AddClusterResponse, retErr error) {
	// Redact the secret from the request
	removeSecret := func(r *lc.AddClusterRequest) *lc.AddClusterRequest {
		copyReq := proto.Clone(r).(*lc.AddClusterRequest)
		copyReq.Secret = ""
		return copyReq
	}
	a.LogReq(ctx, removeSecret(req))
	defer func(start time.Time) { a.pachLogger.Log(ctx, removeSecret(req), nil, retErr, time.Since(start)) }(time.Now())

	// Make sure we have an active license
	if err := a.checkLicenseState(ctx); err != nil {
		return nil, err
	}

	// Validate the request
	if req.Id == "" {
		return nil, errors.New("no id provided for cluster")
	}

	if err := a.validateClusterConfig(ctx, req.Address); err != nil {
		return nil, err
	}

	// Generate the new shared secret for this pachd
	secret := req.Secret
	if secret == "" {
		secret = random.String(30)
	}

	// Register the pachd in the database
	if _, err := a.env.DB.ExecContext(ctx,
		`INSERT INTO license.clusters (id, address, secret, cluster_deployment_id, user_address, is_enterprise_server, version, auth_enabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, req.Id, req.Address, secret, req.ClusterDeploymentId, req.UserAddress, req.EnterpriseServer, "unknown", false); err != nil {
		// throw a unique error if the error is a primary key uniqueness violation
		if dbutil.IsUniqueViolation(err) {
			return nil, lc.ErrDuplicateClusterID
		}
		return nil, errors.Wrapf(err, "unable to register pachd in database")
	}

	return &lc.AddClusterResponse{
		Secret: secret,
	}, nil
}

func stripSecretFromRequest(req *lc.HeartbeatRequest) *lc.HeartbeatRequest {
	r := *req
	r.Secret = ""
	return &r
}

func (a *apiServer) Heartbeat(ctx context.Context, req *lc.HeartbeatRequest) (resp *lc.HeartbeatResponse, retErr error) {
	redactedRequest := stripSecretFromRequest(req)
	a.LogReq(ctx, redactedRequest)
	defer func(start time.Time) { a.pachLogger.Log(ctx, redactedRequest, nil, retErr, time.Since(start)) }(time.Now())

	var count int
	if err := a.env.DB.GetContext(ctx, &count, `SELECT COUNT(*) FROM license.clusters WHERE id=$1 and secret=$2`, req.Id, req.Secret); err != nil {
		return nil, errors.Wrapf(err, "unable to look up cluster in database")
	}

	if count != 1 {
		return nil, lc.ErrInvalidIDOrSecret
	}

	if _, err := a.env.DB.ExecContext(ctx, `UPDATE license.clusters SET version=$1, auth_enabled=$2, client_id=$3, last_heartbeat=NOW() WHERE id=$4`, req.Version, req.AuthEnabled, req.ClientId, req.Id); err != nil {
		return nil, errors.Wrapf(err, "unable to update cluster in database")
	}

	var record ec.LicenseRecord
	if err := a.license.ReadOnly(ctx).Get(licenseRecordKey, &record); err != nil {
		return nil, err
	}

	return &lc.HeartbeatResponse{
		License: &record,
	}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *lc.DeleteAllRequest) (resp *lc.DeleteAllResponse, retErr error) {
	a.LogReq(ctx, req)
	defer func(start time.Time) { a.pachLogger.Log(ctx, req, resp, retErr, time.Since(start)) }(time.Now())

	// TODO: attempt to synchronously deactivate enterprise licensing on every registered cluster

	if _, err := a.env.DB.ExecContext(ctx, `DELETE FROM license.clusters`); err != nil {
		return nil, errors.Wrapf(err, "unable to delete clusters in database")
	}

	if err := dbutil.WithTx(ctx, a.env.DB, func(sqlTx *pachsql.Tx) error {
		err := a.license.ReadWrite(sqlTx).Delete(licenseRecordKey)
		if err != nil && !col.IsErrNotFound(err) {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &lc.DeleteAllResponse{}, nil
}

func (a *apiServer) ListClusters(ctx context.Context, req *lc.ListClustersRequest) (resp *lc.ListClustersResponse, retErr error) {
	a.LogReq(ctx, req)
	defer func(start time.Time) { a.pachLogger.Log(ctx, req, resp, retErr, time.Since(start)) }(time.Now())

	clusters := make([]*lc.ClusterStatus, 0)
	err := a.env.DB.SelectContext(ctx, &clusters, "SELECT id, address, version, auth_enabled, last_heartbeat FROM license.clusters;")
	if err != nil {
		return nil, err
	}

	return &lc.ListClustersResponse{
		Clusters: clusters,
	}, nil
}

func (a *apiServer) DeleteCluster(ctx context.Context, req *lc.DeleteClusterRequest) (resp *lc.DeleteClusterResponse, retErr error) {
	a.LogReq(ctx, req)
	defer func(start time.Time) { a.pachLogger.Log(ctx, req, resp, retErr, time.Since(start)) }(time.Now())

	// TODO: attempt to synchronously deactivate enterprise licensing on the specified cluster

	_, err := a.env.DB.ExecContext(ctx, "DELETE FROM license.clusters WHERE id=$1", req.Id)
	if err != nil {
		return nil, err
	}
	return &lc.DeleteClusterResponse{}, nil
}

func (a *apiServer) UpdateCluster(ctx context.Context, req *lc.UpdateClusterRequest) (resp *lc.UpdateClusterResponse, retErr error) {
	a.LogReq(ctx, req)
	defer func(start time.Time) { a.pachLogger.Log(ctx, req, resp, retErr, time.Since(start)) }(time.Now())

	if req.Address != "" {
		if err := a.validateClusterConfig(ctx, req.Address); err != nil {
			return nil, err
		}
	}

	fieldValues := make(map[string]string)
	fieldValues["address"] = req.Address
	fieldValues["user_address"] = req.UserAddress
	fieldValues["cluster_deployment_id"] = req.ClusterDeploymentId

	var setFields string
	for k, v := range fieldValues {
		if v != "" {
			setFields += fmt.Sprintf(" %s = '%s',", k, v)
		}
	}

	if setFields == "" {
		return nil, errors.New("No cluster fields were provided to the UpdateCluster RPC")
	}

	// trim trailing comma
	setFields = setFields[:len(setFields)-1]

	_, err := a.env.DB.ExecContext(ctx, "UPDATE license.clusters SET "+setFields+"  WHERE id=$1", req.Id)
	if err != nil {
		return nil, err
	}
	return &lc.UpdateClusterResponse{}, nil
}

func (a *apiServer) ListUserClusters(ctx context.Context, req *lc.ListUserClustersRequest) (resp *lc.ListUserClustersResponse, retErr error) {
	a.LogReq(ctx, req)
	defer func(start time.Time) { a.pachLogger.Log(ctx, req, resp, retErr, time.Since(start)) }(time.Now())
	clusters := make([]*lc.UserClusterInfo, 0)
	if err := a.env.DB.SelectContext(ctx, &clusters, `SELECT id, cluster_deployment_id, user_address, is_enterprise_server FROM license.clusters WHERE is_enterprise_server = false`); err != nil {
		return nil, err
	}
	return &lc.ListUserClustersResponse{
		Clusters: clusters,
	}, nil
}
