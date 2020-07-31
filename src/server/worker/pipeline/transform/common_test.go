package transform

import (
	"context"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func defaultPipelineInfo() *pps.PipelineInfo {
	name := "testPipeline"
	return &pps.PipelineInfo{
		Pipeline:     client.NewPipeline(name),
		OutputBranch: "master",
		Transform: &pps.Transform{
			Cmd:        []string{"bash"},
			Stdin:      []string{"cp inputRepo/* out"},
			WorkingDir: client.PPSInputPrefix,
		},
		ParallelismSpec: &pps.ParallelismSpec{
			Constant: 1,
		},
		ResourceRequests: &pps.ResourceSpec{
			Memory: "100M",
			Cpu:    0.5,
		},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "inputRepo",
				Repo:   "inputRepo",
				Branch: "master",
				Glob:   "/*",
			},
		},
		SpecCommit: client.NewCommit(ppsconsts.SpecRepo, name),
	}
}

type testEnv struct {
	*testpachd.RealEnv
	logger *logs.MockLogger
	driver driver.Driver
}

// withTestEnv provides a test env with etcd and pachd instances and connected
// clients, plus a worker driver for performing worker operations.
func withTestEnv(pipelineInfo *pps.PipelineInfo, cb func(*testEnv) error, customConfig ...*serviceenv.PachdFullConfiguration) error {
	return testpachd.WithRealEnv(func(realEnv *testpachd.RealEnv) error {
		logger := logs.NewMockLogger()
		workerDir := filepath.Join(realEnv.Directory, "worker")
		driver, err := driver.NewDriver(
			pipelineInfo,
			realEnv.PachClient,
			realEnv.EtcdClient,
			"/pachyderm_test",
			filepath.Join(workerDir, "hashtrees"),
			workerDir,
			"namespace",
			false,
		)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(realEnv.PachClient.Ctx())
		defer cancel()
		driver = driver.WithContext(ctx)

		env := &testEnv{
			RealEnv: realEnv,
			logger:  logger,
			driver:  driver,
		}

		return cb(env)
	}, customConfig...)
}
