package cmds

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type validationEnv struct{}

func (ve validationEnv) DefaultCPURequest() resource.Quantity {
	return resource.MustParse("1m")
}

func (ve validationEnv) DefaultMemoryRequest() resource.Quantity {
	return resource.MustParse("1G")
}

func (ve validationEnv) DefaultStorageRequest() resource.Quantity {
	return resource.MustParse("1G")
}

func (ve validationEnv) ImagePullSecrets() []string { return nil }

func (ve validationEnv) S3GatewayPort() uint16 { return 1234 }
func (ve validationEnv) PostgresSecretRef(context.Context) (*v1.SecretKeySelector, error) {
	return nil, nil
}
func (ve validationEnv) ImagePullPolicy() string       { return "" }
func (ve validationEnv) WorkerImage() string           { return "workerimg" }
func (ve validationEnv) SidecarImage() string          { return "sidecarimg" }
func (ve validationEnv) StorageRoot() string           { return "/" }
func (ve validationEnv) Namespace() string             { return "pachyderm" }
func (ve validationEnv) StorageBackend() string        { return "foo" }
func (ve validationEnv) PostgresUser() string          { return "dummy-user" }
func (ve validationEnv) PostgresDatabase() string      { return "pgdb" }
func (ve validationEnv) PGBouncerHost() string         { return "localhost" }
func (ve validationEnv) PGBouncerPort() uint16         { return 5432 }
func (ve validationEnv) PeerPort() uint16              { return 1235 }
func (ve validationEnv) LokiHost() string              { return "localhost" }
func (ve validationEnv) LokiPort() uint16              { return 12346 }
func (ve validationEnv) SidecarPort() uint16           { return 1237 }
func (ve validationEnv) InSidecars() bool              { return false }
func (ve validationEnv) GarbageCollectionPercent() int { return 75 }
func (ve validationEnv) SidecarEnvVars(pi *pps.PipelineInfo, ev []v1.EnvVar) []v1.EnvVar {
	return ev
}
func (ve validationEnv) EtcdPrefix() string                  { return "/foobar" }
func (ve validationEnv) PPSWorkerPort() uint16               { return 1238 }
func (ve validationEnv) CommitProgressCounterDisabled() bool { return false }
func (ve validationEnv) LokiLoggingEnabled() bool            { return true }
func (ve validationEnv) GoogleCloudProfilerProject() string  { return "" }
func (ve validationEnv) StorageHostPath() string             { return "/tmp/abcd" }
func (ve validationEnv) TLSSecretName() string               { return "" }
func (ve validationEnv) WorkerSecurityContextsEnabled() bool { return false }
func (ve validationEnv) WorkerUsesRoot() bool                { return false }
