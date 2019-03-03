package serviceenv

type Configuration struct {
	FeatureFlags
	// Ports served by Pachd
	PPSWorkerPort uint16 `env:"PPS_WORKER_GRPC_PORT,default=80"`
	Port          uint16 `env:"PORT,default=650"`
	PProfPort     uint16 `env:"PPROF_PORT,default=651"`
	HTTPPort      uint16 `env:"HTTP_PORT,default=652"`
	PeerPort      uint16 `env:"PEER_PORT,default=653"`

	NumShards             uint64 `env:"NUM_SHARDS,default=32"`
	StorageRoot           string `env:"PACH_ROOT,default=/pach"`
	StorageBackend        string `env:"STORAGE_BACKEND,default="`
	StorageHostPath       string `env:"STORAGE_HOST_PATH,default="`
	EtcdPrefix            string `env:"ETCD_PREFIX,default="`
	PPSEtcdPrefix         string `env:"PPS_ETCD_PREFIX,default=pachyderm_pps"`
	PFSEtcdPrefix         string `env:"PFS_ETCD_PREFIX,default=pachyderm_pfs"`
	AuthEtcdPrefix        string `env:"PACHYDERM_AUTH_ETCD_PREFIX,default=pachyderm_auth"`
	EnterpriseEtcdPrefix  string `env:"PACHYDERM_ENTERPRISE_ETCD_PREFIX,default=pachyderm_enterprise"`
	KubeAddress           string `env:"KUBERNETES_PORT_443_TCP_ADDR,required"`
	EtcdHost              string `env:"ETCD_SERVICE_HOST,required"`
	EtcdPort              string `env:"ETCD_SERVICE_PORT,required"`
	Namespace             string `env:"NAMESPACE,default=default"`
	Metrics               bool   `env:"METRICS,default=true"`
	Init                  bool   `env:"INIT,default=false"`
	BlockCacheBytes       string `env:"BLOCK_CACHE_BYTES,default=1G"`
	PFSCacheSize          string `env:"PFS_CACHE_SIZE,default=0"`
	WorkerImage           string `env:"WORKER_IMAGE,default="`
	WorkerSidecarImage    string `env:"WORKER_SIDECAR_IMAGE,default="`
	WorkerImagePullPolicy string `env:"WORKER_IMAGE_PULL_POLICY,default="`
	LogLevel              string `env:"LOG_LEVEL,default=info"`
	IAMRole               string `env:"IAM_ROLE,default="`
	ImagePullSecret       string `env:"IMAGE_PULL_SECRET,default="`
	NoExposeDockerSocket  bool   `env:"NO_EXPOSE_DOCKER_SOCKET,default=false"`
	ExposeObjectAPI       bool   `env:"EXPOSE_OBJECT_API,default=false"`
	MemoryRequest         string `env:"PACHD_MEMORY_REQUEST,default=1T"`
	WorkerUsesRoot        bool   `env:"WORKER_USES_ROOT,default=true"`
}

type FeatureFlags struct {
	NewHashTree bool `env:"NEW_HASH_TREE,default=false"`
}
