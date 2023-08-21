package v2_8_0

import (
	"context"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
)

func synthesizeClusterDefaults(ctx context.Context, env migrations.Env) error {
	if err := env.LockTables(ctx, "collections.cluster_defaults"); err != nil {
		return errors.EnsureStack(err)
	}
	envMap, err := envMap(os.Environ())
	if err != nil {
		return errors.Wrap(err, "could not convert os.Environ to map")
	}
	cd, err := defaultsFromEnv(envMap)
	if err != nil {
		return err
	}
	js, err := protojson.Marshal(cd)
	if err != nil {
		return errors.Wrap(err, "could not marshal cluster defaults to JSON")
	}
	wrapper := &ppsdb.ClusterDefaultsWrapper{Json: string(js)}
	if err := ppsdb.CollectionsV2_7_0()[0].ReadWrite(env.Tx).Create("", wrapper); err != nil {
		return errors.Wrap(err, "could not create cluster defaults")
	}

	return nil
}

func defaultsFromEnv(envMap map[string]string) (*pps.ClusterDefaults, error) {
	var (
		cd = &pps.ClusterDefaults{
			CreatePipelineRequest: &pps.CreatePipelineRequest{
				ResourceRequests:        &pps.ResourceSpec{},
				SidecarResourceRequests: &pps.ResourceSpec{},
			},
		}
		err error
	)
	if cd.CreatePipelineRequest.ResourceRequests.Memory, err = getQuantityStringFromEnv(envMap, "PIPELINE_DEFAULT_MEMORY_REQUEST", "256Mi"); err != nil {
		return nil, errors.Wrap(err, "could not synthesize pipeline default memory request")
	}
	if cd.CreatePipelineRequest.ResourceRequests.Cpu, err = getFloat32FromEnv(envMap, "PIPELINE_DEFAULT_CPU_REQUEST", 1); err != nil {
		return nil, errors.Wrap(err, "could not synthesize pipeline default CPU request")
	}
	if cd.CreatePipelineRequest.ResourceRequests.Disk, err = getQuantityStringFromEnv(envMap, "PIPELINE_DEFAULT_STORAGE_REQUEST", "1Gi"); err != nil {
		return nil, errors.Wrap(err, "could not synthesize pipeline default disk request")
	}
	if cd.CreatePipelineRequest.SidecarResourceRequests.Memory, err = getQuantityStringFromEnv(envMap, "SIDECAR_DEFAULT_MEMORY_REQUEST", "256Mi"); err != nil {
		return nil, errors.Wrap(err, "could not synthesize sidecar default memory request")
	}
	if cd.CreatePipelineRequest.SidecarResourceRequests.Cpu, err = getFloat32FromEnv(envMap, "SIDECAR_DEFAULT_CPU_REQUEST", 1); err != nil {
		return nil, errors.Wrap(err, "could not synthesize pipeline default CPU request")
	}
	if cd.CreatePipelineRequest.SidecarResourceRequests.Disk, err = getQuantityStringFromEnv(envMap, "SIDECAR_DEFAULT_STORAGE_REQUEST", "1Gi"); err != nil {
		return nil, errors.Wrap(err, "could not synthesize pipeline default disk request")
	}
	return cd, nil
}

var errEmptyVariable = errors.New("empty environment variable")

func getQuantityFromEnv(env map[string]string, name, defaultValue string) (resource.Quantity, error) {
	var (
		q       resource.Quantity
		val, ok = env[name]
		err     error
	)
	if val == "" {
		if ok {
			return q, errors.Wrap(errEmptyVariable, name)
		}
		val = defaultValue
	}
	if q, err = resource.ParseQuantity(val); err != nil {
		return q, errors.Wrapf(err, "could not parse %s (%q) as quantity", name, val)
	}
	if q.Sign() < 0 {
		return q, errors.Errorf("negative quantity %v for %s (%s)", q, name, val)
	}
	return q, nil
}

func getQuantityStringFromEnv(env map[string]string, name, defaultValue string) (string, error) {
	q, err := getQuantityFromEnv(env, name, defaultValue)
	if err != nil {
		return "", errors.Wrap(err, "could not get quantity")
	}
	return q.String(), nil
}

func getFloat32FromEnv(env map[string]string, name string, defaultValue float32) (float32, error) {
	q, err := getQuantityFromEnv(env, name, resource.NewMilliQuantity(int64(defaultValue*1000), resource.DecimalSI).String())
	if err != nil {
		return 0, errors.Wrap(err, "could not get quantity")
	}
	return float32(q.AsApproximateFloat64()), nil

}

// envMap converts a list of NAME=VAL pairs (as returned by os.Environ) to a map
// from names to values.  Later values overwrite earlier ones.
func envMap(environ []string) (map[string]string, error) {
	var env = make(map[string]string)
	for _, e := range environ {
		ee := strings.SplitN(e, "=", 2)
		if len(ee) == 1 {
			return nil, errors.Errorf("invalid environment member %s", e)
		}
		env[ee[0]] = ee[1]
	}
	return env, nil
}
