package transform

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func defaultPipelineInfo() *pps.PipelineInfo {
	name := "testPipeline"
	return &pps.PipelineInfo{
		Pipeline:     client.NewPipeline(name),
		OutputBranch: "master",
		Transform: &pps.Transform{
			Cmd: []string{"cp", path.Join("pfs", "inputRepo", "file"), "pfs/out/file"},
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
				Repo:   "inputRepo",
				Branch: "master",
				Glob:   "/*",
			},
		},
		SpecCommit: client.NewCommit(ppsconsts.SpecRepo, name),
	}
}

// A Driver instance that provides a pachClient
type mockDriver struct {
	*driver.MockDriver
	pachClient *client.APIClient
}

func newMockDriver(env *tu.RealEnv, pipelineInfo *pps.PipelineInfo) *mockDriver {
	return &mockDriver{
		MockDriver: driver.NewMockDriver(env.EtcdClient, &driver.MockOptions{
			NumWorkers:   1,
			PipelineInfo: pipelineInfo,
			HashtreePath: path.Join(env.Directory, "hashtrees"),
		}),
		pachClient: env.PachClient,
	}
}

func (md *mockDriver) PachClient() *client.APIClient {
	return md.pachClient
}

func (md *mockDriver) WithCtx(ctx context.Context) driver.Driver {
	result := &mockDriver{}
	*result = *md
	result.MockDriver = md.MockDriver.WithCtx(ctx).(*driver.MockDriver)
	result.pachClient = md.PachClient().WithCtx(ctx)
	return result
}

type testEnv struct {
	*tu.RealEnv
	logger *logs.MockLogger
	driver *mockDriver
}

func withTestEnv(pipelineInfo *pps.PipelineInfo, cb func(*testEnv) error) error {
	return tu.WithRealEnv(func(realEnv *tu.RealEnv) error {
		ctx, cancel := context.WithCancel(realEnv.PachClient.Ctx())
		defer cancel()

		mockDriver := newMockDriver(realEnv, pipelineInfo).WithCtx(ctx).(*mockDriver)
		mockLogger := logs.NewMockLogger()

		env := &testEnv{
			RealEnv: realEnv,
			logger:  mockLogger,
			driver:  mockDriver,
		}

		return cb(env)
	})
}
