package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type execDriver struct {
	mu        sync.RWMutex
	pipelines map[string]context.CancelFunc
	crashCh   chan watch.Event
	logger    *logrus.Logger
}

func newExecDriver(logger *logrus.Logger) *execDriver {
	return &execDriver{
		pipelines: make(map[string]context.CancelFunc),
		crashCh:   make(chan watch.Event, 1),
		logger:    logger,
	}
}

func (ed *execDriver) CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error {
	ctx, cf := context.WithCancel(ctx)
	ed.mu.Lock()
	ed.pipelines[pi.Pipeline.Name] = cf
	ed.mu.Unlock()
	logFile := "/Users/alon/.pachyderm/logs/workers/log.txt"
	os.MkdirAll("/Users/alon/.pachyderm/logs/workers", 0700)
	os.Remove(logFile)
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				cmd := ed.makeExec(ctx, pi, logFile)
				fmt.Printf("executing cmd: %v\n", cmd.String())
				if err := cmd.Run(); err == nil {
					break
				} else {
					logrus.New().Errorf("ERROR EXECUTING %q, : err: %v", cmd, err)
					ed.crashCh <- watch.Event{Type: watch.Error}
				}
			}
		}
	}()
	return nil
}

func (ed *execDriver) makeExec(ctx context.Context, pi *pps.PipelineInfo, logFile string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "go", "run", "/Users/alon/workspace/pachyderm/src/server/cmd/worker/main.go")
	cmd.Env = append([]string{
		fmt.Sprintf("PPS_PIPELINE_NAME=%s", pi.Pipeline.Name),
		fmt.Sprintf("PPS_SPEC_COMMIT=%s", pi.SpecCommit.ID),
		fmt.Sprintf("PPS_WORKER_IP=%s", "127.0.0.1"),
		fmt.Sprintf("PPS_POD_NAME=%s", pi.Pipeline.Name),
		"PEER_PORT=1650",
		fmt.Sprintf("PACH_ROOT=/Users/alon/.pachyderm/data/workers/%s", pi.Pipeline.Name),

		fmt.Sprintf("PPS_WORKER_GRPC_PORT=1081"),
	},
		os.Environ()...,
	)
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	cmd.Stdout = f
	cmd.Stderr = f
	return cmd
}

func (ed *execDriver) DeletePipelineResources(ctx context.Context, pipeline string) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	if cf, ok := ed.pipelines[pipeline]; ok {
		cf()
	}
	delete(ed.pipelines, pipeline)
	return nil
}

func (ed *execDriver) ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationControllerList, error) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()
	if _, ok := ed.pipelines[pi.Pipeline.Name]; !ok {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{},
		}, errors.New("rc for pipeline not found")
	} else {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: ppsutil.PipelineRcName(pi.Pipeline.Name, pi.Version),
						Annotations: map[string]string{
							pachVersionAnnotation:        version.PrettyVersion(),
							hashedAuthTokenAnnotation:    hashAuthToken(pi.AuthToken),
							pipelineVersionAnnotation:    strconv.FormatUint(pi.Version, 10),
							pipelineSpecCommitAnnotation: pi.SpecCommit.ID,
						},
					},
				},
			},
		}, nil
	}
}

func (*execDriver) UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error {
	return nil
}

func (*execDriver) ListReplicationControllers(ctx context.Context) (*v1.ReplicationControllerList, error) {
	return nil, nil
}

func (ed *execDriver) WatchPipelinePods(ctx context.Context) (<-chan watch.Event, context.CancelFunc, error) {
	return ed.crashCh, func() {}, nil
}
