package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/random"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
)

const (
	licenseRecordKey             = "license"
	localhostEnterpriseClusterId = "localhost"
)

type apiServer struct {
	lc.UnimplementedAPIServer

	env *Env
	// license is the database record where we store the active enterprise license
	license col.PostgresCollection
}

// New returns an implementation of license.APIServer, and a function that bootstraps the license server via environment.
func New(env *Env) (*apiServer, error) {
	s := &apiServer{
		env:     env,
		license: licenseCollection(env.DB, env.Listener),
	}
	return s, nil
}

func (a *apiServer) EnvBootstrap(ctx context.Context) error {
	if a.env.Config.LicenseKey == "" && a.env.Config.EnterpriseSecret == "" {
		return nil
	}
	if a.env.Config.LicenseKey == "" || a.env.Config.EnterpriseSecret == "" {
		return errors.New("License server failed to bootstrap via environment; Either both or neither of LICENSE_KEY and ENTERPRISE_SECRET must be set.")
	}
	log.Info(ctx, "Started to configure license server via environment")
	localhostPeerAddr := "grpc://localhost:" + fmt.Sprint(a.env.Config.PeerPort)
	if err := func() error {
		ctx = auth.AsInternalUser(ctx, "license-server")
		_, err := a.activate(ctx, &lc.ActivateRequest{ActivationCode: a.env.Config.LicenseKey})
		if err != nil {
			return errors.Wrapf(err, "activate the license service")
		}
		_, err = a.AddCluster(ctx, &lc.AddClusterRequest{
			Id:               localhostEnterpriseClusterId,
			Address:          localhostPeerAddr,
			UserAddress:      localhostPeerAddr,
			Secret:           a.env.Config.EnterpriseSecret,
			EnterpriseServer: true,
		})
		if err != nil {
			if errors.As(err, lc.ErrDuplicateClusterID) {
				_, err = a.UpdateCluster(ctx, &lc.UpdateClusterRequest{
					Id:          localhostEnterpriseClusterId,
					Address:     localhostPeerAddr,
					UserAddress: localhostPeerAddr,
					Secret:      a.env.Config.EnterpriseSecret,
				})
				if err != nil {
					return errors.Wrapf(err, "update localhost cluster in the license service")
				}
			} else {
				return errors.Wrapf(err, "add localhost cluster in the license service")
			}
		}
		_, err = a.env.EnterpriseServer.Activate(ctx, &ec.ActivateRequest{
			Id:            localhostEnterpriseClusterId,
			LicenseServer: localhostPeerAddr,
			Secret:        a.env.Config.EnterpriseSecret})
		if err != nil {
			return errors.Wrapf(errors.EnsureStack(err), "activate localhost cluster in the enterprise service")
		}
		return nil
	}(); err != nil {
		return errors.Errorf("bootstrap license service from the environment: %v", err)
	}
	log.Info(ctx, "Successfully configured license server via environment")
	return nil
}

// Activate implements the Activate RPC
func (a *apiServer) Activate(ctx context.Context, req *lc.ActivateRequest) (*lc.ActivateResponse, error) {
	if a.env.Config.LicenseKey != "" {
		return nil, errors.New("license.Activate() is disabled when the license key is configured via environment")
	}
	return a.activate(ctx, req)
}

func (a *apiServer) activate(ctx context.Context, req *lc.ActivateRequest) (resp *lc.ActivateResponse, retErr error) {
	// Validate the activation code
	expiration, err := license.Validate(req.ActivationCode)
	if err != nil {
		return nil, errors.Wrapf(err, "error validating activation code")
	}
	// Allow request to override expiration in the activation code, for testing
	if req.Expires != nil {
		customExpiration := req.Expires.AsTime()
		if err == nil && expiration.After(customExpiration) {
			expiration = customExpiration
		}
	}
	expirationProto := timestamppb.New(expiration)

	newRecord := &ec.LicenseRecord{
		ActivationCode: req.ActivationCode,
		Expires:        expirationProto,
	}

	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		return errors.EnsureStack(a.license.ReadWrite(sqlTx).Put(licenseRecordKey, newRecord))
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
	return a.getLicenseRecord(ctx)
}

func (a *apiServer) getLicenseRecord(ctx context.Context) (*lc.GetActivationCodeResponse, error) {
	var record ec.LicenseRecord
	if err := a.license.ReadOnly(ctx).Get(licenseRecordKey, &record); err != nil {
		if col.IsErrNotFound(err) {
			return &lc.GetActivationCodeResponse{State: ec.State_NONE}, nil
		}
		return nil, errors.EnsureStack(err)
	}

	expiration := record.Expires.AsTime()
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
		return errors.Errorf("enterprise license is not valid - %v", record.State)
	}
	return nil
}

func (a *apiServer) validateClusterConfig(ctx context.Context, address string) error {
	if address == "" {
		return errors.New("no address provided for cluster")
	}
	return nil
}

// AddCluster registers a new pachd with this license server. Each pachd is configured with a shared secret
// which is used to authenticate to the license server when heartbeating.
func (a *apiServer) AddCluster(ctx context.Context, req *lc.AddClusterRequest) (resp *lc.AddClusterResponse, retErr error) {
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

func (a *apiServer) Heartbeat(ctx context.Context, req *lc.HeartbeatRequest) (resp *lc.HeartbeatResponse, retErr error) {
	var count int
	if err := a.env.DB.GetContext(ctx, &count, `SELECT COUNT(*) FROM license.clusters WHERE id=$1 and secret=$2`, req.Id, req.Secret); err != nil {
		return nil, errors.Wrapf(err, "unable to look up cluster in database")
	}

	if count != 1 {
		return nil, errors.Wrapf(lc.ErrInvalidIDOrSecret, "got %d, want 1 for cluster id %q", count, req.Id)
	}

	if _, err := a.env.DB.ExecContext(ctx, `UPDATE license.clusters SET version=$1, auth_enabled=$2, client_id=$3, last_heartbeat=NOW() WHERE id=$4`, req.Version, req.AuthEnabled, req.ClientId, req.Id); err != nil {
		return nil, errors.Wrapf(err, "unable to update cluster in database")
	}

	var record ec.LicenseRecord
	if err := a.license.ReadOnly(ctx).Get(licenseRecordKey, &record); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &lc.HeartbeatResponse{
		License: &record,
	}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *lc.DeleteAllRequest) (resp *lc.DeleteAllResponse, retErr error) {
	// TODO: attempt to synchronously deactivate enterprise licensing on every registered cluster

	if _, err := a.env.DB.ExecContext(ctx, `DELETE FROM license.clusters`); err != nil {
		return nil, errors.Wrapf(err, "unable to delete clusters in database")
	}

	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		err := a.license.ReadWrite(sqlTx).Delete(licenseRecordKey)
		if err != nil && !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &lc.DeleteAllResponse{}, nil
}

type dbClusterStatus struct {
	Id            string     `db:"id"`
	Address       string     `db:"address"`
	Version       string     `db:"version"`
	AuthEnabled   bool       `db:"auth_enabled"`
	ClientId      string     `db:"client_id"`
	LastHeartbeat *time.Time `db:"last_heartbeat"`
	CreatedAt     *time.Time `db:"created_at"`
}

func (cs dbClusterStatus) ToProto() *lc.ClusterStatus {
	return &lc.ClusterStatus{
		Id:            cs.Id,
		Address:       cs.Address,
		Version:       cs.Version,
		AuthEnabled:   cs.AuthEnabled,
		ClientId:      cs.ClientId,
		LastHeartbeat: protoutil.MustTimestampFromPointer(cs.LastHeartbeat),
		CreatedAt:     protoutil.MustTimestampFromPointer(cs.CreatedAt),
	}
}

func (a *apiServer) ListClusters(ctx context.Context, req *lc.ListClustersRequest) (resp *lc.ListClustersResponse, retErr error) {
	var clusters []dbClusterStatus
	err := a.env.DB.SelectContext(ctx, &clusters, "SELECT id, address, version, auth_enabled, last_heartbeat FROM license.clusters")
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	result := make([]*lc.ClusterStatus, len(clusters))
	for i, cs := range clusters {
		result[i] = cs.ToProto()
	}
	return &lc.ListClustersResponse{
		Clusters: result,
	}, nil
}

func (a *apiServer) DeleteCluster(ctx context.Context, req *lc.DeleteClusterRequest) (resp *lc.DeleteClusterResponse, retErr error) {
	// TODO: attempt to synchronously deactivate enterprise licensing on the specified cluster

	_, err := a.env.DB.ExecContext(ctx, "DELETE FROM license.clusters WHERE id=$1", req.Id)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &lc.DeleteClusterResponse{}, nil
}

func (a *apiServer) UpdateCluster(ctx context.Context, req *lc.UpdateClusterRequest) (resp *lc.UpdateClusterResponse, retErr error) {
	if req.Address != "" {
		if err := a.validateClusterConfig(ctx, req.Address); err != nil {
			return nil, err
		}
	}

	fieldValues := make(map[string]string)
	fieldValues["address"] = req.Address
	fieldValues["user_address"] = req.UserAddress
	fieldValues["cluster_deployment_id"] = req.ClusterDeploymentId
	fieldValues["secret"] = req.Secret

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
		return nil, errors.EnsureStack(err)
	}
	return &lc.UpdateClusterResponse{}, nil
}

func (a *apiServer) ListUserClusters(ctx context.Context, req *lc.ListUserClustersRequest) (resp *lc.ListUserClustersResponse, retErr error) {
	clusters := make([]*lc.UserClusterInfo, 0)
	if err := a.env.DB.SelectContext(ctx, &clusters, `SELECT id, cluster_deployment_id AS ClusterDeploymentId, user_address AS Address, is_enterprise_server AS EnterpriseServer FROM license.clusters WHERE is_enterprise_server = false`); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &lc.ListUserClustersResponse{
		Clusters: clusters,
	}, nil
}
