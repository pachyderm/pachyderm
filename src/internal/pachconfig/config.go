// Package pachconfig contains the configuration models for Pachyderm.
// The various models are parsed from environment variables when pachd starts.
// This package should be at the bottom of the dependency graph.
package pachconfig

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// Configuration is the generic configuration structure used to access configuration fields.
type Configuration struct {
	*GlobalConfiguration
	*PachdSpecificConfiguration
	*WorkerSpecificConfiguration
	*EnterpriseSpecificConfiguration
}

// GlobalConfiguration contains the global configuration.  Note that the logger is initialized
// before service environment setup, but uses environment variables.  See the documentation for the
// src/internal/log package for those.
type GlobalConfiguration struct {
	FeatureFlags
	PostgresConfiguration

	EtcdHost               string `env:"ETCD_SERVICE_HOST,required"`
	EtcdPort               string `env:"ETCD_SERVICE_PORT,required"`
	PPSWorkerPort          uint16 `env:"PPS_WORKER_GRPC_PORT,default=1080"`
	PPSWorkerPreprocessing bool   `env:"PPS_WORKER_PREPROCESSING,default=false"`
	Port                   uint16 `env:"PORT,default=1650"`
	PrometheusPort         uint16 `env:"PROMETHEUS_PORT,default=1656"`
	PeerPort               uint16 `env:"PEER_PORT,default=1653"`
	S3GatewayPort          uint16 `env:"S3GATEWAY_PORT,default=1600"`
	DownloadPort           uint16 `env:"DOWNLOAD_PORT,default=1659"`
	PPSEtcdPrefix          string `env:"PPS_ETCD_PREFIX,default=pachyderm_pps"`
	Namespace              string `env:"PACH_NAMESPACE,default=default"`
	StorageRoot            string `env:"PACH_ROOT,default=/pach"`
	GCPercent              int    `env:"GC_PERCENT,default=100"`
	LokiHost               string `env:"LOKI_SERVICE_HOST"`
	LokiPort               string `env:"LOKI_SERVICE_PORT"`
	OidcPort               uint16 `env:"OIDC_PORT,default=1657"`
	IsPachw                bool   `env:"IS_PACHW,default=false"`
	PachwMinReplicas       int    `env:"PACHW_MIN_REPLICAS"`
	PachwMaxReplicas       int    `env:"PACHW_MAX_REPLICAS,default=1"`
	PachdServiceHost       string `env:"PACHD_SERVICE_HOST"`
	PachdServicePort       string `env:"PACHD_SERVICE_PORT"`

	EtcdPrefix           string `env:"ETCD_PREFIX,default="`
	DeploymentID         string `env:"CLUSTER_DEPLOYMENT_ID,default="`
	EnterpriseEtcdPrefix string `env:"PACHYDERM_ENTERPRISE_ETCD_PREFIX,default=pachyderm_enterprise"`
	Metrics              bool   `env:"METRICS,default=true"`
	MetricsEndpoint      string `env:"METRICS_ENDPOINT,default="`

	// SessionDurationMinutes it how long auth tokens are valid for, defaults to 30 days (30 * 24 * 60)
	SessionDurationMinutes int `env:"SESSION_DURATION_MINUTES,default=43200"`

	IdentityServerDatabase string `env:"IDENTITY_SERVER_DATABASE,default=dex"`

	// PPSSpecCommitID and PPSPipelineName are only set for workers and sidecar
	// pachd instances. Because both pachd and worker need to know the spec commit
	// (the worker so that it can avoid jobs for other versions of the same pipelines
	// and the sidecar so that it can serve the S3 gateway) it's stored in the
	// GlobalConfiguration, but it isn't set in a cluster's main pachd containers.
	PPSSpecCommitID string `env:"PPS_SPEC_COMMIT"`
	// The name of the project that this worker belongs to.
	PPSProjectName string `env:"PPS_PROJECT_NAME"`
	// The name of the pipeline that this worker belongs to.
	PPSPipelineName string `env:"PPS_PIPELINE_NAME"`

	// If set to the name of a GCP project, enable GCP-specific continuous profiling and send
	// profiles to that project: https://cloud.google.com/profiler/docs.  Requires that pachd
	// has google application credentials (through environment variables or workload identity),
	// and that the service account associated with the credentials has 'cloudprofiler.agent' on
	// the target project.  If set on a pachd pod, propagates to workers and sidecars (which
	// also need permission).
	GoogleCloudProfilerProject string `env:"GOOGLE_CLOUD_PROFILER_PROJECT"`

	// The number of concurrent requests that the PPS Master can make against kubernetes
	PPSMaxConcurrentK8sRequests int `env:"PPS_MAX_CONCURRENT_K8S_REQUESTS,default=10"`

	// These are automatically injected into pachd by Kubernetes so that Go's GC can be tuned to
	// take advantage of the memory made available to the container.  They should not be set by
	// users manually; use GOMEMLIMIT directly instead.
	K8sMemoryLimit   float64 `env:"K8S_MEMORY_LIMIT,default=0"`
	K8sMemoryRequest float64 `env:"K8S_MEMORY_REQUEST,default=0"`

	// Users tend to have a bad experience when they request 0 resources from k8s.  These are
	// the defaults for piplines that don't supply any requests or limits.  (As soon as you
	// supply and request or limit, even an empty request or limit, then these are all ignored.)
	PipelineDefaultMemoryRequest  resource.Quantity `env:"PIPELINE_DEFAULT_MEMORY_REQUEST,default=256Mi"`
	PipelineDefaultCPURequest     resource.Quantity `env:"PIPELINE_DEFAULT_CPU_REQUEST,default=1"`
	PipelineDefaultStorageRequest resource.Quantity `env:"PIPELINE_DEFAULT_STORAGE_REQUEST,default=1Gi"`

	SidecarDefaultMemoryRequest  resource.Quantity `env:"SIDECAR_DEFAULT_MEMORY_REQUEST,default=256Mi"`
	SidecarDefaultCPURequest     resource.Quantity `env:"SIDECAR_DEFAULT_CPU_REQUEST,default=1"`
	SidecarDefaultStorageRequest resource.Quantity `env:"SIDECAR_DEFAULT_STORAGE_REQUEST,default=1Gi"`
}

func (GlobalConfiguration) isPachConfig() {}

// PostgresConfiguration configures postgres and pg-bouncer.
type PostgresConfiguration struct {
	PostgresSSL                    string `env:"POSTGRES_SSL,default=disable"`
	PostgresHost                   string `env:"POSTGRES_HOST"`
	PostgresPort                   int    `env:"POSTGRES_PORT"`
	PostgresDBName                 string `env:"POSTGRES_DATABASE"`
	PostgresUser                   string `env:"POSTGRES_USER"`
	PostgresPassword               string `env:"POSTGRES_PASSWORD"`
	PostgresMaxOpenConns           int    `env:"POSTGRES_MAX_OPEN_CONNS,default=10"`
	PostgresMaxIdleConns           int    `env:"POSTGRES_MAX_IDLE_CONNS,default=10"`
	PostgresConnMaxLifetimeSeconds int    `env:"POSTGRES_CONN_MAX_LIFETIME_SECONDS,default=0"`
	PostgresConnMaxIdleSeconds     int    `env:"POSTGRES_CONN_MAX_IDLE_SECONDS,default=0"`
	PostgresQueryLogging           bool   `env:"POSTGRES_QUERY_LOGGING,default=false"`

	PGBouncerMaxOpenConns int    `env:"PG_BOUNCER_MAX_OPEN_CONNS,default=10"`
	PGBouncerMaxIdleConns int    `env:"PG_BOUNCER_MAX_IDLE_CONNS,default=10"`
	PGBouncerHost         string `env:"PG_BOUNCER_HOST,required"`
	PGBouncerPort         int    `env:"PG_BOUNCER_PORT,required"`
}

// PachdFullConfiguration contains the full pachd configuration.
type PachdFullConfiguration struct {
	GlobalConfiguration
	PachdSpecificConfiguration
	EnterpriseSpecificConfiguration
}

func (PachdFullConfiguration) isPachConfig() {}

// PachdSpecificConfiguration contains the pachd specific configuration.
type PachdSpecificConfiguration struct {
	StorageConfiguration
	StorageBackend             string `env:"STORAGE_BACKEND,required"`
	StorageURL                 string `env:"STORAGE_URL,default="`
	StorageHostPath            string `env:"STORAGE_HOST_PATH,default="`
	PFSEtcdPrefix              string `env:"PFS_ETCD_PREFIX,default=pachyderm_pfs"`
	KubeAddress                string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	Init                       bool   `env:"INIT,default=false"`
	WorkerImage                string `env:"WORKER_IMAGE,default="`
	WorkerSidecarImage         string `env:"WORKER_SIDECAR_IMAGE,default="`
	WorkerImagePullPolicy      string `env:"WORKER_IMAGE_PULL_POLICY,default="`
	ImagePullSecrets           string `env:"IMAGE_PULL_SECRETS,default="`
	MemoryRequest              string `env:"PACHD_MEMORY_REQUEST,default=1T"`
	WorkerUsesRoot             bool   `env:"WORKER_USES_ROOT,default=false"`
	EnablePreflightChecks      bool   `env:"ENABLE_PREFLIGHT_CHECKS,default=true"`
	RequireCriticalServersOnly bool   `env:"REQUIRE_CRITICAL_SERVERS_ONLY,default=false"`
	// TODO: Merge this with the worker specific pod name (PPS_POD_NAME) into a global configuration pod name.
	PachdPodName                 string `env:"PACHD_POD_NAME,required"`
	EnableWorkerSecurityContexts bool   `env:"ENABLE_WORKER_SECURITY_CONTEXTS,default=true"`
	TLSCertSecretName            string `env:"TLS_CERT_SECRET_NAME,default="`

	// Now that Pachyderm has HTTP endpoints, we need to be able to link users to the HTTP
	// endpoint.  These two variables handle that; ProxyHost for the user-accessible location of
	// the proxy, and ProxyTLS for whether or not to use https:// for generated URLs.
	ProxyHost string `env:"PACHYDERM_PUBLIC_HOST,default="`
	ProxyTLS  bool   `env:"PACHYDERM_PUBLIC_TLS,default=false"`
}

// EnterpriseServerConfiguration contains the full configuration for an enterprise server
type EnterpriseServerConfiguration struct {
	GlobalConfiguration
	EnterpriseSpecificConfiguration
}

// EnterpriseSpecificConfiguration contains the configuration required for enterprise features
type EnterpriseSpecificConfiguration struct {
	ActivateAuth              bool   `env:"ACTIVATE_AUTH,default=false"`
	AuthRootToken             string `env:"AUTH_ROOT_TOKEN,default="`
	AuthConfig                string `env:"AUTH_CONFIG,default="`
	AuthClientSecret          string `env:"AUTH_CLIENT_SECRET,default="`
	AuthClusterRoleBindings   string `env:"AUTH_CLUSTER_RBAC,default="`
	LicenseKey                string `env:"LICENSE_KEY,default="`
	EnterpriseSecret          string `env:"ENTERPRISE_SECRET,default="`
	EnterpriseMember          bool   `env:"ENTERPRISE_MEMBER,default=false"`
	EnterpriseServerAddress   string `env:"ENTERPRISE_SERVER_ADDRESS,default="`
	EnterpriseServerToken     string `env:"ENTERPRISE_SERVER_TOKEN,default="`
	EnterpriseMemberConfig    string `env:"ENTERPRISE_MEMBER_CONFIG,default="`
	IdentityConfig            string `env:"IDP_CONFIG,default="`
	IdentityConnectors        string `env:"IDP_CONNECTORS,default="`
	IdentityClients           string `env:"IDP_CLIENTS,default="`
	IdentityAdditionalClients string `env:"IDP_ADDITIONAL_CLIENTS,default="`
	TrustedPeers              string `env:"TRUSTED_PEERS,default="`
	ConsoleOAuthID            string `env:"CONSOLE_OAUTH_ID,default="`
	ConsoleOAuthSecret        string `env:"CONSOLE_OAUTH_SECRET,default="`
	DeterminedOAuthID         string `env:"DETERMINED_OAUTH_ID,default="`
	DeterminedOAuthSecret     string `env:"DETERMINED_OAUTH_SECRET,default="`
	// Determined integration configuration
	DeterminedUsername string `env:"DETERMINED_USERNAME,default="`
	DeterminedPassword string `env:"DETERMINED_PASSWORD,default="`
	DeterminedURL      string `env:"DETERMINED_API_URL,default="`
	DeterminedTLS      bool   `env:"DETERMINED_TLS,default=false"`
}

// StorageConfiguration contains the storage configuration.
type StorageConfiguration struct {
	StorageMemoryThreshold               int64 `env:"STORAGE_MEMORY_THRESHOLD"`
	StorageCompactionShardSizeThreshold  int64 `env:"STORAGE_COMPACTION_SHARD_SIZE_THRESHOLD"`
	StorageCompactionShardCountThreshold int64 `env:"STORAGE_COMPACTION_SHARD_COUNT_THRESHOLD"`
	StorageLevelFactor                   int64 `env:"STORAGE_LEVEL_FACTOR"`
	StorageUploadConcurrencyLimit        int   `env:"STORAGE_UPLOAD_CONCURRENCY_LIMIT,default=100"`
	StorageDownloadConcurrencyLimit      int   `env:"STORAGE_DOWNLOAD_CONCURRENCY_LIMIT,default=100"`
	StoragePutFileConcurrencyLimit       int   `env:"STORAGE_PUT_FILE_CONCURRENCY_LIMIT,default=100"`
	StorageGCPeriod                      int64 `env:"STORAGE_GC_PERIOD,default=60"`
	StorageChunkGCPeriod                 int64 `env:"STORAGE_CHUNK_GC_PERIOD,default=60"`
	StorageCompactionMaxFanIn            int   `env:"STORAGE_COMPACTION_MAX_FANIN,default=10"`
	StorageFileSetsMaxOpen               int   `env:"STORAGE_FILESETS_MAX_OPEN,default=50"`
	StorageDiskCacheSize                 int   `env:"STORAGE_DISK_CACHE_SIZE,default=100"`
	StorageMemoryCacheSize               int   `env:"STORAGE_MEMORY_CACHE_SIZE,default=100"`
}

// WorkerFullConfiguration contains the full worker configuration.
type WorkerFullConfiguration struct {
	GlobalConfiguration
	WorkerSpecificConfiguration
}

// WorkerSpecificConfiguration contains the worker specific configuration.
type WorkerSpecificConfiguration struct {
	// Worker gets its own IP here, via the k8s downward API. It then writes that
	// IP back to etcd so that pachd can discover it
	PPSWorkerIP string `env:"PPS_WORKER_IP,required"`
	// The name of this pod
	PodName string `env:"PPS_POD_NAME,required"`
}

// FeatureFlags contains the configuration for feature flags.  XXX: if you're
// adding a new feature flag then you need to make sure it gets propagated to
// the workers and their sidecars, this should be done in:
// src/server/pps/server/worker_rc.go in the workerPodSpec func.
type FeatureFlags struct {
	DisableCommitProgressCounter bool `env:"DISABLE_COMMIT_PROGRESS_COUNTER,default=false"`
	LokiLogging                  bool `env:"LOKI_LOGGING,default=false"`
	IdentityServerEnabled        bool `env:"IDENTITY_SERVER_ENABLED,default=false"`
}

// PachdPreflightConfiguration is configuration for the preflight checks.
type PachdPreflightConfiguration struct {
	PostgresConfiguration
}

func (PachdPreflightConfiguration) isPachConfig() {}

// PachdRestoreSnapshotConfiguration is configuration for the restoring from a snapshot.
type PachdRestoreSnapshotConfiguration struct {
	PostgresConfiguration
}

func (p PachdRestoreSnapshotConfiguration) isPachConfig() {}

// NewConfiguration creates a generic configuration from a specific type of configuration.
func NewConfiguration(config any) *Configuration {
	configuration := &Configuration{}
	switch v := config.(type) {
	case *GlobalConfiguration:
		configuration.GlobalConfiguration = v
		return configuration
	case *PachdFullConfiguration:
		configuration.GlobalConfiguration = &v.GlobalConfiguration
		configuration.PachdSpecificConfiguration = &v.PachdSpecificConfiguration
		configuration.EnterpriseSpecificConfiguration = &v.EnterpriseSpecificConfiguration
		return configuration
	case *WorkerFullConfiguration:
		configuration.GlobalConfiguration = &v.GlobalConfiguration
		configuration.WorkerSpecificConfiguration = &v.WorkerSpecificConfiguration
		return configuration
	case *EnterpriseServerConfiguration:
		configuration.GlobalConfiguration = &v.GlobalConfiguration
		configuration.EnterpriseSpecificConfiguration = &v.EnterpriseSpecificConfiguration
		return configuration
	case *PachdPreflightConfiguration:
		configuration.GlobalConfiguration = &GlobalConfiguration{
			PostgresConfiguration: v.PostgresConfiguration,
		}
		return configuration
	case *PachdRestoreSnapshotConfiguration:
		configuration.GlobalConfiguration = &GlobalConfiguration{
			PostgresConfiguration: v.PostgresConfiguration,
		}
		return configuration
	default:
		return nil
	}
}

// EmptyConfig is only used in pachctl-doc.
//
// TODO: remove
type EmptyConfig struct{}

func (EmptyConfig) isPachConfig() {}

type AnyConfig interface {
	isPachConfig()
}
