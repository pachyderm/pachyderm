package serviceenv

// Configuration is the generic configuration structure used to access configuration fields.
type Configuration struct {
	*GlobalConfiguration
	*PachdSpecificConfiguration
	*WorkerSpecificConfiguration
}

// GlobalConfiguration contains the global configuration.
type GlobalConfiguration struct {
	FeatureFlags
	EtcdHost                       string `env:"ETCD_SERVICE_HOST,required"`
	EtcdPort                       string `env:"ETCD_SERVICE_PORT,required"`
	PPSWorkerPort                  uint16 `env:"PPS_WORKER_GRPC_PORT,default=1080"`
	Port                           uint16 `env:"PORT,default=1650"`
	PeerPort                       uint16 `env:"PEER_PORT,default=1653"`
	S3GatewayPort                  uint16 `env:"S3GATEWAY_PORT,default=1600"`
	PPSEtcdPrefix                  string `env:"PPS_ETCD_PREFIX,default=pachyderm_pps"`
	Namespace                      string `env:"PACH_NAMESPACE,default=default"`
	StorageRoot                    string `env:"PACH_ROOT,default=/pach"`
	GCPercent                      int    `env:"GC_PERCENT,default=50"`
	LokiHost                       string `env:"LOKI_SERVICE_HOST"`
	LokiPort                       string `env:"LOKI_SERVICE_PORT"`
	OidcPort                       uint16 `env:"OIDC_PORT,default=1657"`
	PostgresHost                   string `env:"POSTGRES_HOST"`
	PostgresPort                   int    `env:"POSTGRES_PORT"`
	PostgresSSL                    string `env:"POSTGRES_SSL,default=disable"`
	PostgresDBName                 string `env:"POSTGRES_DATABASE_NAME"`
	PostgresUser                   string `env:"POSTGRES_USER,default=postgres"`
	PostgresPassword               string `env:"POSTGRES_PASSWORD"`
	PostgresMaxOpenConns           int    `env:"POSTGRES_MAX_OPEN_CONNS"`
	PostgresMaxIdleConns           int    `env:"POSTGRES_MAX_IDLE_CONNS"`
	PostgresConnMaxLifetimeSeconds int    `env:"POSTGRES_CONN_MAX_LIFETIME_SECONDS"`
	PostgresConnMaxIdleSeconds     int    `env:"POSTGRES_CONN_MAX_IDLE_SECONDS"`
	PachdServiceHost               string `env:"PACHD_SERVICE_HOST"`
	PachdServicePort               string `env:"PACHD_SERVICE_PORT"`

	EtcdPrefix           string `env:"ETCD_PREFIX,default="`
	DeploymentID         string `env:"CLUSTER_DEPLOYMENT_ID,default="`
	LogLevel             string `env:"LOG_LEVEL,default=info"`
	EnterpriseEtcdPrefix string `env:"PACHYDERM_ENTERPRISE_ETCD_PREFIX,default=pachyderm_enterprise"`
	Metrics              bool   `env:"METRICS,default=true"`
	MetricsEndpoint      string `env:"METRICS_ENDPOINT,default="`

	// SessionDurationMinutes it how long auth tokens are valid for, defaults to 30 days (30 * 24 * 60)
	SessionDurationMinutes int `env:"SESSION_DURATION_MINUTES,default=43200"`

	IdentityServerDatabase string `env:"IDENTITY_SERVER_DATABASE,default=dex"`
	IdentityServerUser     string `env:"IDENTITY_SERVER_USER,default=postgres"`
	IdentityServerPassword string `env:"IDENTITY_SERVER_PASSWORD"`

	// PPSSpecCommitID and PPSPipelineName are only set for workers and sidecar
	// pachd instances. Because both pachd and worker need to know the spec commit
	// (the worker so that it can avoid jobs for other versions of the same pipelines
	// and the sidecar so that it can serve the S3 gateway) it's stored in the
	// GlobalConfiguration, but it isn't set in a cluster's main pachd containers.
	PPSSpecCommitID string `env:"PPS_SPEC_COMMIT"`
	// The name of the pipeline that this worker belongs to
	PPSPipelineName string `env:"PPS_PIPELINE_NAME"`
}

// PachdFullConfiguration contains the full pachd configuration.
type PachdFullConfiguration struct {
	GlobalConfiguration
	PachdSpecificConfiguration
}

// PachdSpecificConfiguration contains the pachd specific configuration.
type PachdSpecificConfiguration struct {
	StorageConfiguration
	StorageBackend             string `env:"STORAGE_BACKEND,required"`
	StorageHostPath            string `env:"STORAGE_HOST_PATH,default="`
	PFSEtcdPrefix              string `env:"PFS_ETCD_PREFIX,default=pachyderm_pfs"`
	KubeAddress                string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	Init                       bool   `env:"INIT,default=false"`
	WorkerImage                string `env:"WORKER_IMAGE,default="`
	WorkerSidecarImage         string `env:"WORKER_SIDECAR_IMAGE,default="`
	WorkerImagePullPolicy      string `env:"WORKER_IMAGE_PULL_POLICY,default="`
	ImagePullSecret            string `env:"IMAGE_PULL_SECRET,default="`
	MemoryRequest              string `env:"PACHD_MEMORY_REQUEST,default=1T"`
	WorkerUsesRoot             bool   `env:"WORKER_USES_ROOT,default=true"`
	RequireCriticalServersOnly bool   `env:"REQUIRE_CRITICAL_SERVERS_ONLY,default=false"`
	// TODO: Merge this with the worker specific pod name (PPS_POD_NAME) into a global configuration pod name.
	PachdPodName string `env:"PACHD_POD_NAME,required"`
}

// StorageConfiguration contains the storage configuration.
type StorageConfiguration struct {
	StorageMemoryThreshold         int64  `env:"STORAGE_MEMORY_THRESHOLD"`
	StorageShardThreshold          int64  `env:"STORAGE_SHARD_THRESHOLD"`
	StorageLevelFactor             int64  `env:"STORAGE_LEVEL_FACTOR"`
	StorageUploadConcurrencyLimit  int    `env:"STORAGE_UPLOAD_CONCURRENCY_LIMIT,default=100"`
	StoragePutFileConcurrencyLimit int    `env:"STORAGE_PUT_FILE_CONCURRENCY_LIMIT,default=100"`
	StorageGCPolling               string `env:"STORAGE_GC_POLLING"`
	StorageGCTimeout               string `env:"STORAGE_GC_TIMEOUT"`
	StorageCompactionMaxFanIn      int    `env:"STORAGE_COMPACTION_MAX_FANIN,default=10"`
	StorageFileSetsMaxOpen         int    `env:"STORAGE_FILESETS_MAX_OPEN,default=50"`
	StorageDiskCacheSize           int    `env:"STORAGE_DISK_CACHE_SIZE,default=100"`
	StorageMemoryCacheSize         int    `env:"STORAGE_MEMORY_CACHE_SIZE,default=100"`
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

// NewConfiguration creates a generic configuration from a specific type of configuration.
func NewConfiguration(config interface{}) *Configuration {
	configuration := &Configuration{}
	switch v := config.(type) {
	case *GlobalConfiguration:
		configuration.GlobalConfiguration = v
		return configuration
	case *PachdFullConfiguration:
		configuration.GlobalConfiguration = &v.GlobalConfiguration
		configuration.PachdSpecificConfiguration = &v.PachdSpecificConfiguration
		return configuration
	case *WorkerFullConfiguration:
		configuration.GlobalConfiguration = &v.GlobalConfiguration
		configuration.WorkerSpecificConfiguration = &v.WorkerSpecificConfiguration
		return configuration
	default:
		return nil
	}
}
