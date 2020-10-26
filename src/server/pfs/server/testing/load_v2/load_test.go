package load

import (
	"fmt"
	"testing"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil/random"
	yaml "gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	msg := random.SeedRand()
	require.NoError(t, testLoad(), msg)
}

func testLoad() error {
	return testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		load := defaultCommitsSpec()
		loadYAML, err := yaml.Marshal(load)
		if err != nil {
			return err
		}
		fmt.Println(string(loadYAML))
		c := env.PachClient
		repo := "test"
		if err := c.CreateRepo(repo); err != nil {
			return err
		}
		return Commits(c, repo, "master", load)
	}, newPachdConfig())
}

func newPachdConfig() *serviceenv.PachdFullConfiguration {
	config := &serviceenv.PachdFullConfiguration{}
	config.StorageV2 = true
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelZeroSize = units.MB
	config.StorageGCPolling = "30s"
	config.StorageCompactionMaxFanIn = 50
	return config
}

func defaultCommitsSpec() *CommitsSpec {
	return &CommitsSpec{
		Count:           5,
		OperationsSpecs: []*OperationsSpec{defaultOperationsSpec()},
		ValidatorSpec:   &ValidatorSpec{},
	}
}

func defaultOperationsSpec() *OperationsSpec {
	return &OperationsSpec{
		Count: 5,
		FuzzOperationSpecs: []*FuzzOperationSpec{
			&FuzzOperationSpec{
				OperationSpec: &OperationSpec{
					&PutTarSpec{
						FilesSpec: defaultFilesSpec(),
					},
				},
				Prob: 1.0,
			},
		},
	}
}

func defaultFilesSpec() *FilesSpec {
	return &FilesSpec{
		Count: 5,
		FuzzFileSpecs: []*FuzzFileSpec{
			&FuzzFileSpec{
				FileSpec: &FileSpec{
					RandomFileSpec: &RandomFileSpec{
						FuzzSizeSpecs: defaultFuzzSizeSpecs(),
					},
				},
				Prob: 1.0,
			},
		},
	}
}

func defaultFuzzSizeSpecs() []*FuzzSizeSpec {
	return []*FuzzSizeSpec{
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 1 * units.KB,
				Max: 10 * units.KB,
			},
			Prob: 0.3,
		},
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 10 * units.KB,
				Max: 100 * units.KB,
			},
			Prob: 0.3,
		},
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 1 * units.MB,
				Max: 10 * units.MB,
			},
			Prob: 0.3,
		},
		&FuzzSizeSpec{
			SizeSpec: &SizeSpec{
				Min: 10 * units.MB,
				Max: 100 * units.MB,
			},
			Prob: 0.1,
		},
	}
}
