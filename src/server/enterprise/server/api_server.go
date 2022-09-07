package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/keycache"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	internalauth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
	logrus "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	// enterpriseTokenKey is the constant key we use that maps to an Enterprise
	// token that a user has given us. This is what we check to know if a
	// Pachyderm cluster supports enterprise features
	enterpriseTokenKey = "token"
	configKey          = "config"

	heartbeatFrequency = time.Hour
	heartbeatTimeout   = time.Minute

	updatedAtFieldName   = "pachyderm.com/updatedAt"
	restartedAtFieldName = "kubectl.kubernetes.io/restartedAt"
)

type apiServer struct {
	env *Env

	enterpriseTokenCache *keycache.Cache

	// enterpriseTokenCol is a collection containing the enterprise license state
	enterpriseTokenCol col.EtcdCollection

	// configCol is a collection containing the license server configuration
	configCol col.PostgresCollection
}

// NewEnterpriseServer returns an implementation of ec.APIServer.
func NewEnterpriseServer(env *Env, heartbeat bool) (*apiServer, error) {
	defaultEnterpriseRecord := &ec.EnterpriseRecord{}
	enterpriseTokenCol := col.NewEtcdCollection(
		env.EtcdClient,
		env.EtcdPrefix,
		nil,
		&ec.EnterpriseRecord{},
		nil,
		nil,
	)

	s := &apiServer{
		env:                  env,
		enterpriseTokenCache: keycache.NewCache(env.BackgroundContext, enterpriseTokenCol.ReadOnly(env.BackgroundContext), enterpriseTokenKey, defaultEnterpriseRecord),
		enterpriseTokenCol:   enterpriseTokenCol,
		configCol:            EnterpriseConfigCollection(env.DB, env.Listener),
	}
	go s.enterpriseTokenCache.Watch()

	if heartbeat {
		go s.heartbeatRoutine()
	}
	return s, nil
}

func (a *apiServer) EnvBootstrap(ctx context.Context) error {
	if !a.env.Config.EnterpriseMember {
		return nil
	}
	a.env.Logger.Info("Started to configure enterprise member cluster via environment")
	var cluster *lc.AddClusterRequest
	if err := yaml.Unmarshal([]byte(a.env.Config.EnterpriseMemberConfig), &cluster); err != nil {
		return errors.Wrapf(err, "unmarshal enterprise cluster %q", a.env.Config.EnterpriseMemberConfig)
	}
	cluster.ClusterDeploymentId = a.env.Config.DeploymentID
	cluster.Secret = a.env.Config.EnterpriseSecret
	es, err := client.NewFromURI(a.env.Config.EnterpriseServerAddress)
	if err != nil {
		return errors.Wrap(err, "connect to enterprise server")
	}
	es.SetAuthToken(a.env.Config.EnterpriseServerToken)
	if _, err := es.License.AddCluster(es.Ctx(), cluster); err != nil {
		if !lc.IsErrDuplicateClusterID(err) {
			return errors.Wrap(err, "add enterprise cluster")
		}
		if _, err = es.License.UpdateCluster(es.Ctx(), &lc.UpdateClusterRequest{
			Id:                  cluster.Id,
			Address:             cluster.Address,
			UserAddress:         cluster.UserAddress,
			ClusterDeploymentId: cluster.ClusterDeploymentId,
			Secret:              cluster.Secret,
		}); err != nil {
			return errors.Wrap(err, "update enterprise cluster")
		}
	}
	ctx = internalauth.AsInternalUser(ctx, auth.InternalPrefix+"enterprise-service")
	if _, err = a.Activate(ctx, &ec.ActivateRequest{
		LicenseServer: a.env.Config.EnterpriseServerAddress,
		Id:            cluster.Id,
		Secret:        cluster.Secret,
	}); err != nil {
		return errors.Wrap(err, "activate enterprise service")
	}
	a.env.Logger.Info("Successfully configured enterprise member cluster via environment")
	return nil
}

// heartbeatRoutine should  be run in a goroutine and attempts to heartbeat to the license service.
// If the attempt fails and the license server is configured it logs the error.
func (a *apiServer) heartbeatRoutine() {
	heartbeat := func() {
		ctx, cancel := context.WithTimeout(context.Background(), heartbeatTimeout)
		defer cancel()
		if err := a.heartbeatIfConfigured(ctx); err != nil && !lc.IsErrNotActivated(err) {
			logrus.WithError(err).Error("enterprise license heartbeat process failed")
		}
	}
	heartbeat()
	for range time.Tick(heartbeatFrequency) {
		heartbeat()
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
		return errors.EnsureStack(err)
	}

	// If we can't reach the license server, we assume our license is still valid.
	// However, if we reach the license server and the cluster ID has been deleted then we should
	// deactivate the license on this server.
	resp, err := a.heartbeatToServer(ctx, config.LicenseServer, config.Id, config.Secret)
	if err != nil {
		if lc.IsErrInvalidIDOrSecret(err) {
			logrus.WithError(err).Error("enterprise license heartbeat had invalid id or secret, disabling enterprise")
			_, err = col.NewSTM(ctx, a.env.EtcdClient, func(stm col.STM) error {
				e := a.enterpriseTokenCol.ReadWrite(stm)
				err := e.Put(enterpriseTokenKey, &ec.EnterpriseRecord{
					LastHeartbeat:   types.TimestampNow(),
					HeartbeatFailed: true,
				})
				return errors.EnsureStack(err)
			})
			return err
		}
		return err
	}

	_, err = col.NewSTM(ctx, a.env.EtcdClient, func(stm col.STM) error {
		e := a.enterpriseTokenCol.ReadWrite(stm)
		err := e.Put(enterpriseTokenKey, &ec.EnterpriseRecord{
			LastHeartbeat:   types.TimestampNow(),
			License:         resp.License,
			HeartbeatFailed: false,
		})
		return errors.EnsureStack(err)
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

	var clientID string
	authEnabled := true
	config, err := a.env.AuthServer.GetConfiguration(ctx, &auth.GetConfigurationRequest{})
	if err != nil && auth.IsErrNotActivated(err) {
		authEnabled = false
	} else if err != nil {
		return nil, errors.EnsureStack(err)
	} else {
		clientID = config.Configuration.ClientID
	}

	pachClient, err := client.NewFromURI(licenseServer)
	if err != nil {
		return nil, err
	}

	res, err := pachClient.License.Heartbeat(ctx, &lc.HeartbeatRequest{
		Id:          id,
		Secret:      secret,
		Version:     versionResp,
		AuthEnabled: authEnabled,
		ClientId:    clientID,
	})
	return res, errors.EnsureStack(err)
}

// Heartbeat implements the Heartbeat RPC. It exists mostly to test the heartbeat logic
func (a *apiServer) Heartbeat(ctx context.Context, req *ec.HeartbeatRequest) (resp *ec.HeartbeatResponse, retErr error) {
	if err := a.heartbeatIfConfigured(ctx); err != nil {
		return nil, err
	}
	return &ec.HeartbeatResponse{}, nil
}

// Activate implements the Activate RPC
func (a *apiServer) Activate(ctx context.Context, req *ec.ActivateRequest) (resp *ec.ActivateResponse, retErr error) {
	// must not activate while paused
	if a.env.mode == PausedMode {
		return nil, errors.New("cannot activate paused cluster; unpause first")
	}
	// Try to heartbeat before persisting the configuration
	heartbeatResp, err := a.heartbeatToServer(ctx, req.LicenseServer, req.Id, req.Secret)
	if err != nil {
		return nil, err
	}

	record := &ec.EnterpriseRecord{License: heartbeatResp.License}

	// If the test heartbeat succeeded, write the state and config to etcd
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txCtx *txncontext.TransactionContext) error {
		if err := a.configCol.ReadWrite(txCtx.SqlTx).Put(configKey, &ec.EnterpriseConfig{
			LicenseServer: req.LicenseServer,
			Id:            req.Id,
			Secret:        req.Secret,
		}); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := col.NewSTM(ctx, a.env.EtcdClient, func(stm col.STM) error {
		return errors.EnsureStack(a.enterpriseTokenCol.ReadWrite(stm).Put(enterpriseTokenKey, record))
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
			return nil, errors.EnsureStack(err)
		}

		resp.ActivationCode = base64.StdEncoding.EncodeToString(activationCodeStr)
	}

	return resp, nil
}

// GetActivationCode returns the current state of the cluster's Pachyderm Enterprise key (ACTIVE, EXPIRED, or NONE), including the enterprise activation code
func (a *apiServer) GetActivationCode(ctx context.Context, req *ec.GetActivationCodeRequest) (resp *ec.GetActivationCodeResponse, retErr error) {
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
	// must not deactivate while paused
	if a.env.mode == PausedMode {
		return nil, errors.New("cannot deactivate paused cluster; unpause first")
	}
	if _, err := col.NewSTM(ctx, a.env.EtcdClient, func(stm col.STM) error {
		err := a.enterpriseTokenCol.ReadWrite(stm).Delete(enterpriseTokenKey)
		if err != nil && !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txCtx *txncontext.TransactionContext) error {
		err := a.configCol.ReadWrite(txCtx.SqlTx).Delete(configKey)
		if err != nil && !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
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

// Pause sets the cluster to paused mode, restarting all pachds in a paused
// status.  If they are already paused, it is a no-op.
func (a *apiServer) Pause(ctx context.Context, req *ec.PauseRequest) (resp *ec.PauseResponse, retErr error) {
	if a.env.mode == UnpausableMode {
		return nil, errors.Errorf("this pachd instance is not in a pausable mode")
	}
	if err := a.rollPachd(ctx, true); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &ec.PauseResponse{}, nil
}

// rollPachd changes pachds from paused to unpaused, or unpaused to paused.  It
// does this by creating a ConfigMap with a MODE value, then updating the pachd
// deployment spec’s template’s annotations with a new value, which causes
// Kubernetes to terminate existing pods and start new ones.
//
// Since the ConfigMap is not managed by Helm, it will persist through Helm
// operations: a cluster which is upgraded by Helm in a paused state will thus
// remain paused.
//
// If the ConfigMap is already in the desired state, no changes are made; this
// ensures that PauseStatus can do its checking appropriately.
//
// There is a special case in main to handle the case where there is no
// ConfigMap and the '$(MODE)' reference is verbatim.
func (a *apiServer) rollPachd(ctx context.Context, paused bool) error {
	var (
		kc        = a.env.getKubeClient()
		namespace = a.env.namespace
	)
	cc := kc.CoreV1().ConfigMaps(namespace)
	c, err := cc.Get(ctx, "pachd-config", metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		// if the ConfigMap is not found, then the cluster is by definition unpaused
		if !paused {
			return nil
		}
		// Since pachd-config is not managed by Helm, it may not exist.
		c = newPachdConfigMap(namespace, a.env.unpausedMode)
		c.Annotations[updatedAtFieldName] = time.Now().Format(time.RFC3339Nano)
		c.Data["MODE"] = "paused"
		if _, err := cc.Create(ctx, c, metav1.CreateOptions{}); err != nil {
			return errors.Errorf("could not create configmap: %w", err)
		}
	} else if err != nil {
		return errors.Errorf("could not get configmap pachd-config: %w", err)
	} else {
		// Curiously, Get can return no error and an empty configmap!
		// If this happens, need to set the name and namespace.
		if c.ObjectMeta.Name == "" {
			c = newPachdConfigMap(namespace, a.env.unpausedMode)
		}
		if c.Data == nil {
			c.Data = make(map[string]string)
		}
		if paused && c.Data["MODE"] == "paused" {
			// Short-circuit and do not update if configmap
			// is already in the correct state.  This keeps
			// the updatedAt semantics correct when checking
			// the pause status.
			return nil
		} else if !paused && c.Data["MODE"] != "paused" {
			// Short-circuit and do not update if configmap
			// is already in the correct state.  This keeps
			// the updatedAt semantics correct when checking
			// the pause status.
			return nil
		}
		if paused {
			c.Data["MODE"] = "paused"
		} else {
			c.Data["MODE"] = a.env.unpausedMode
		}
		if c.Annotations == nil {
			c.Annotations = make(map[string]string)
		}
		c.Annotations[updatedAtFieldName] = time.Now().Format(time.RFC3339Nano)
		if _, err := cc.Update(ctx, c, metav1.UpdateOptions{}); err != nil {
			return errors.Errorf("could not update configmap: %w", err)
		}
	}

	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dd := kc.AppsV1().Deployments(namespace)
		d, err := dd.Get(ctx, "pachd", metav1.GetOptions{})
		if err != nil {
			return err //nolint:wrapcheck
		}
		// Updating the spec rolls the deployment, killing each pod and causing
		// a new one to start.
		d.Spec.Template.Annotations[restartedAtFieldName] = time.Now().Format(time.RFC3339Nano)
		if _, err := dd.Update(ctx, d, metav1.UpdateOptions{}); err != nil {
			return err //nolint:wrapcheck
		}
		return nil
	}); err != nil {
		return errors.Errorf("could not updated pachd deployment: %v", err)
	}

	return nil
}

func newPachdConfigMap(namespace, unpausedMode string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pachd-config",
			Namespace:   namespace,
			Annotations: make(map[string]string),
		},
		Data: map[string]string{"MODE": unpausedMode},
	}
}

func scaleDownWorkers(ctx context.Context, kc kubernetes.Interface, namespace string) error {
	rc := kc.CoreV1().ReplicationControllers(namespace)
	ww, err := rc.List(ctx, metav1.ListOptions{
		LabelSelector: "suite=pachyderm,component=worker",
	})
	if err != nil {
		return errors.Errorf("could not list workers: %v", err)
	}
	for _, w := range ww.Items {
		if _, err := rc.UpdateScale(ctx, w.GetName(), &autoscalingv1.Scale{
			Spec: autoscalingv1.ScaleSpec{
				Replicas: 0,
			},
		}, metav1.UpdateOptions{
			FieldManager: "enterprise-server",
		}); err != nil {
			return errors.Errorf("could not scale down %s: %v", w.GetName(), err)
		}
	}
	return nil
}

func (a *apiServer) Unpause(ctx context.Context, req *ec.UnpauseRequest) (resp *ec.UnpauseResponse, retErr error) {
	if a.env.mode == UnpausableMode {
		return nil, errors.Errorf("this pachd instance is not in an unpausable mode")
	}
	if err := a.rollPachd(ctx, false); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &ec.UnpauseResponse{}, nil
}

// PauseStatus checks the pause status of a cluster.  It does this by checking
// first for the pachd-config ConfigMap; if it does not exist, then the cluster
// is by default unpaused.  Next it looks for the “pachyderm.com/updatedAt”
// field; if that field does not exist then this must be an empty config map and
// the cluster is by default paused.  If it does exist, though, it is remembered.
//
// Finally it cycles through all pachd pods, checking that they were created
// before the updateAt value; if some are before & some after (or at the exact
// same instant), then the cluster is considered partially paused; if all pachd
// pods were created before the ConfigMap was updated then it has not taken
// effect and the cluster is in the reverse state; if all pods were created
// after the ConfigMap was updated then it has taken effect and the cluster is
// in the indicated state.
func (a *apiServer) PauseStatus(ctx context.Context, req *ec.PauseStatusRequest) (resp *ec.PauseStatusResponse, retErr error) {
	kc := a.env.getKubeClient()
	cc := kc.CoreV1().ConfigMaps(a.env.namespace)
	c, err := cc.Get(ctx, "pachd-config", metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		// If there is no configmap, then the pachd pods must be
		// unpaused.
		return &ec.PauseStatusResponse{
			Status: ec.PauseStatusResponse_UNPAUSED,
		}, nil
	} else if err != nil {
		return nil, errors.Errorf("could not get configmap: %v", err)
	}
	if c.Annotations[updatedAtFieldName] == "" {
		// If there is no updatedAt, then then the pachd pods must be in
		// their default, unpaused, state.
		return &ec.PauseStatusResponse{
			Status: ec.PauseStatusResponse_UNPAUSED,
		}, nil
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, c.Annotations[updatedAtFieldName])
	if err != nil {
		return nil, errors.Errorf("could not parse update time %v: %v", c.Annotations[updatedAtFieldName], err)
	}

	pods := kc.CoreV1().Pods(a.env.namespace)
	pp, err := pods.List(ctx, metav1.ListOptions{
		LabelSelector: "app=pachd",
	})
	if err != nil {
		return nil, errors.Errorf("could not get pachd pods: %v", err)
	}
	var sawAfter, sawBefore bool
	for _, p := range pp.Items {
		// We cannot rely on p.CreationTimestamp.Time, because it only
		// has one-second resolution.  Instead, we store our own
		// nanosecond-resolution timestamp.
		//
		// If the pod does not have the Kubernetes restartedAt
		// annotation, then it must have existed before the configMap
		// was updated.
		if p.Annotations[restartedAtFieldName] == "" {
			sawBefore = true
			continue
		}
		restartedAt, err := time.Parse(time.RFC3339Nano, p.Annotations[restartedAtFieldName])
		if err != nil {
			return nil, errors.Errorf("could not parse restarted time %v: %v", p.Annotations[restartedAtFieldName], err)
		}
		if restartedAt.Before(updatedAt) {
			sawBefore = true
		} else {
			sawAfter = true
		}
	}
	var status ec.PauseStatusResponse_PauseStatus
	switch {
	case sawBefore && sawAfter:
		status = ec.PauseStatusResponse_PARTIALLY_PAUSED
	case sawBefore:
		// nothing has cycled yet; configmap has yet to take effect
		if c.Data["MODE"] == "paused" {
			status = ec.PauseStatusResponse_UNPAUSED
		} else {
			status = ec.PauseStatusResponse_PAUSED
		}
	case sawAfter:
		// everything has cycled; configmap has taken effect
		if c.Data["MODE"] == "paused" {
			status = ec.PauseStatusResponse_PAUSED
		} else {
			status = ec.PauseStatusResponse_UNPAUSED
		}
	default:
		return nil, errors.Errorf("no paused or unpaused pachds found")
	}
	return &ec.PauseStatusResponse{
		Status: status,
	}, nil
}
