package v2_8_0

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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
		return nil, errors.Wrap(err, "could not synthesize sidecar default CPU request")
	}
	if cd.CreatePipelineRequest.SidecarResourceRequests.Disk, err = getQuantityStringFromEnv(envMap, "SIDECAR_DEFAULT_STORAGE_REQUEST", "1Gi"); err != nil {
		return nil, errors.Wrap(err, "could not synthesize sidecar default disk request")
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

func versionKey(p *pps.Pipeline, version uint64) string {
	// zero pad in case we want to sort
	return fmt.Sprintf("%s@%08d", p, version)
}

var pipelinesVersionIndex = &index{
	Name: "version",
	Extract: func(val proto.Message) string {
		info := val.(*pps.PipelineInfo)
		return versionKey(info.Pipeline, info.Version)
	},
}

func pipelinesNameKey(p *pps.Pipeline) string {
	return p.String()
}

var pipelinesNameIndex = &index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return pipelinesNameKey(val.(*pps.PipelineInfo).Pipeline)
	},
}

var pipelinesIndexes = []*index{
	pipelinesVersionIndex,
	pipelinesNameIndex,
}

func pipelineCommitKey(commit *pfs.Commit) (string, error) {
	if commit.Repo.Type != pfs.SpecRepoType {
		return "", errors.Errorf("commit %s is not from a spec repo", commit)
	}
	if projectName := commit.Repo.Project.GetName(); projectName != "" {
		return fmt.Sprintf("%s/%s@%s", projectName, commit.Repo.Name, commit.Id), nil
	}
	return fmt.Sprintf("%s@%s", commit.Repo.Name, commit.Id), nil
}

func withKeyGen(gen func(interface{}) (string, error)) colOption {
	return func(c *postgresCollection) {
		c.keyGen = gen
	}
}

func withKeyCheck(check func(string) error) colOption {
	return func(c *postgresCollection) {
		c.keyCheck = check
	}
}

func parsePipelineKey(key string) (projectName, pipelineName, id string, err error) {
	parts := strings.Split(key, "@")
	if len(parts) != 2 || !uuid.IsUUIDWithoutDashes(parts[1]) {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>", key)
	}
	id = parts[1]
	parts = strings.Split(parts[0], "/")
	if len(parts) == 0 {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>")
	}
	pipelineName = parts[len(parts)-1]
	if len(parts) == 1 {
		return
	}
	projectName = strings.Join(parts[0:len(parts)-1], "/")
	return
}

func synthesizeSpec(pi *pps.PipelineInfo) error {
	if pi == nil {
		return errors.New("nil PipelineInfo")
	}
	// create an initial user and effective spec equal to what would have been previously used
	spec := ppsutil.PipelineReqFromInfo(pi)
	js, err := protojson.Marshal(spec)
	if err != nil {
		return errors.Wrapf(err, "could not marshal CreatePipelineRequest as JSON")
	}
	if pi.EffectiveSpecJson == "" {
		pi.EffectiveSpecJson = string(js)
	}
	if pi.UserSpecJson == "" {
		pi.UserSpecJson = string(js)
	}
	return nil
}

func synthesizeSpecs(ctx context.Context, env migrations.Env) error {
	var pipelineInfo = new(pps.PipelineInfo)
	if err := migratePostgreSQLCollection(ctx, env.Tx, "pipelines", pipelinesIndexes, pipelineInfo, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if err = synthesizeSpec(pipelineInfo); err != nil {
			return "", nil, err
		}
		if newKey, err = pipelineCommitKey(pipelineInfo.SpecCommit); err != nil {
			return
		}
		return newKey, pipelineInfo, nil

	},
		withKeyGen(func(key interface{}) (string, error) {
			if commit, ok := key.(*pfs.Commit); ok {
				return pipelineCommitKey(commit)
			}
			return "", errors.New("must provide a spec commit")
		}),
		withKeyCheck(func(key string) error {
			_, _, _, err := parsePipelineKey(key)
			return err
		}),
	); err != nil {
		return errors.Wrap(err, "could not migrate jobs")
	}
	return nil
}
