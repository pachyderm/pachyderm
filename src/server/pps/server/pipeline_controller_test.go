package server

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	logrus "github.com/sirupsen/logrus"
)

func withPPSMaster(t *testing.T, iDriver InfraDriver, stateMgr PipelineStateManager, test func()) {
	etcdEnv := testetcd.NewEnv(t)
	env := Env{
		BackgroundContext: context.Background(),
		Logger:            logrus.StandardLogger(),
		DB:                nil,
		EtcdClient:        etcdEnv.EtcdClient,
		KubeClient:        nil,
		AuthServer:        nil,
		PFSServer:         nil,
		TxnEnv:            nil,
		GetPachClient:     nil, // will need to fix this bit
		Config:            newConfig(t),
	}
	ctx, cf := context.WithCancel(context.Background())
	t.Cleanup(func() { cf() })
	master := newMaster(ctx, env, env.EtcdPrefix, iDriver, stateMgr)
	go master.run()
	test()
}

func TestPipelineController(t *testing.T) {
	stateMgr := newMockStateManager()
	infraDriver := newMockInfraDriver()
	name := tu.UniqueString("TestPipelineController_")
	withPPSMaster(t, infraDriver, stateMgr, func() {
		stateMgr.upsertPipeline(&pps.PipelineInfo{
			Pipeline: client.NewPipeline(name),
			State:    pps.PipelineState_PIPELINE_STARTING,
			Details:  &pps.PipelineInfo_Details{},
		})
		require.NoErrorWithinT(t, 30*time.Second, func() error {
			return backoff.Retry(func() error {
				if stateMgr.pipelines[name].State == pps.PipelineState_PIPELINE_RUNNING {
					return nil
				}
				return errors.New("change hasn't reflected")
			}, backoff.NewTestingBackOff())
		})
	})
}
