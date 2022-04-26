package server

import (
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type InfraDriver interface {
	// Creates a pipeline's services, secrets, and replication controllers.
	CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error
	// Deletes a pipeline's services, secrets, and replication controllers.
	// NOTE: It doesn't return a stepError, leaving retry behavior to the caller
	DeletePipelineResources(ctx context.Context, pipeline string) error
	ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationControllerList, error)
	// UpdateReplicationController intends to server {scaleUp,scaleDown}Pipeline.
	// It includes all of the logic for writing an updated RC spec to kubernetes,
	// and updating/retrying if k8s rejects the write. It presents a strange API,
	// since the the RC being updated is already available to the caller, but update()
	// may be called muliple times if the k8s write fails. It may be helpful to think
	// of the 'old' rc passed to update() as mutable.
	UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error
	ListReplicationControllers(ctx context.Context) (*v1.ReplicationControllerList, error)
	WatchPipelinePods(ctx context.Context) (<-chan watch.Event, func(), error)
}

type mockInfraOp int32

const (
	mockInfraOp_CREATE = 0
	mockInfraOp_DELETE = 1
	mockInfraOp_READ   = 2
)

type mockInfraDriver struct {
	rcs           map[string]v1.ReplicationController // indexed by pipeline name
	calls         map[string]map[mockInfraOp]int      // indexed by pipeline name
	elapsedScales map[string][]int32                  // indexed by pipeline name
}

func newMockInfraDriver() *mockInfraDriver {
	mid := &mockInfraDriver{
		rcs:           make(map[string]v1.ReplicationController),
		calls:         make(map[string]map[mockInfraOp]int),
		elapsedScales: make(map[string][]int32),
	}
	return mid
}

func (mid *mockInfraDriver) CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error {
	mid.rcs[pi.Pipeline.Name] = *mid.makeRC(pi)
	mid.incCall(pi.Pipeline.Name, mockInfraOp_CREATE)
	if _, ok := mid.elapsedScales[pi.Pipeline.Name]; !ok {
		mid.elapsedScales[pi.Pipeline.Name] = make([]int32, 0)
	}
	return nil
}

func (mid *mockInfraDriver) DeletePipelineResources(ctx context.Context, pipeline string) error {
	if _, ok := mid.rcs[pipeline]; ok {
		delete(mid.rcs, pipeline)
		mid.incCall(pipeline, mockInfraOp_DELETE)
	}
	return nil
}

func (mid *mockInfraDriver) ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationControllerList, error) {
	mid.incCall(pi.Pipeline.Name, mockInfraOp_READ)
	if rc, ok := mid.rcs[pi.Pipeline.Name]; !ok {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{},
		}, errors.New("rc for pipeline not found")
	} else {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{rc},
		}, nil
	}
}

func (mid *mockInfraDriver) UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error {
	rc := old.DeepCopy()
	if update(rc) {
		name := rc.ObjectMeta.Labels[pipelineNameLabel]
		mid.elapsedScales[name] = append(mid.elapsedScales[name], *rc.Spec.Replicas)
		mid.writeRC(rc)
	}
	return nil
}

func (mid *mockInfraDriver) ListReplicationControllers(ctx context.Context) (*v1.ReplicationControllerList, error) {
	items := make([]v1.ReplicationController, 0)
	for _, rc := range mid.rcs {
		items = append(items, rc)
	}
	return &v1.ReplicationControllerList{
		Items: items,
	}, nil
}

// TODO(acohen4): complete
func (mid *mockInfraDriver) WatchPipelinePods(ctx context.Context) (<-chan watch.Event, func(), error) {
	ch := make(chan watch.Event)
	return ch, func() {}, nil
}

////////////////////////////////////
// -------- Mock Helpers -------- //
////////////////////////////////////

func (mid *mockInfraDriver) resetRCs() {
	mid.rcs = make(map[string]v1.ReplicationController)
}

func (mid *mockInfraDriver) makeRC(pi *pps.PipelineInfo) *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name: ppsutil.PipelineRcName(pi.Pipeline.Name, pi.Version),
			Annotations: map[string]string{
				pipelineVersionAnnotation:    strconv.FormatUint(pi.Version, 10),
				pipelineSpecCommitAnnotation: pi.SpecCommit.ID,
				hashedAuthTokenAnnotation:    hashAuthToken(pi.AuthToken),
				pachVersionAnnotation:        version.PrettyVersion(),
			},
			Labels: map[string]string{
				pipelineNameLabel: pi.Pipeline.Name,
			},
		},
	}
}

func (mid *mockInfraDriver) writeRC(rc *v1.ReplicationController) {
	name := rc.ObjectMeta.Labels[pipelineNameLabel]
	mid.rcs[name] = *rc
}

func (mid *mockInfraDriver) incCall(name string, op mockInfraOp) {
	if ops, ok := mid.calls[name]; ok {
		ops[op] += 1
	} else {
		mid.calls[name] = map[mockInfraOp]int{op: 1}
	}
}
