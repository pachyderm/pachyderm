package server

import (
	"context"
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/version"
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

type MockInfraOp int32

const (
	MockInfraOp_CREATE = 0
	MockInfraOp_DELETE = 1
	MockInfraOp_READ   = 2
)

type MockInfraDriver struct {
	Rcs          map[string]v1.ReplicationController // indexed by pipeline name
	Calls        map[string]map[MockInfraOp]int      // indexed by pipeline name
	ScaleHistory map[string][]int32                  // indexed by pipeline name
}

func NewMockInfraDriver() *MockInfraDriver {
	d := &MockInfraDriver{
		Rcs:          make(map[string]v1.ReplicationController),
		Calls:        make(map[string]map[MockInfraOp]int),
		ScaleHistory: make(map[string][]int32),
	}
	return d
}

func (d *MockInfraDriver) CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error {
	d.Rcs[pi.Pipeline.Name] = *d.MakeRC(pi)
	d.IncCall(pi.Pipeline.Name, MockInfraOp_CREATE)
	if _, ok := d.ScaleHistory[pi.Pipeline.Name]; !ok {
		d.ScaleHistory[pi.Pipeline.Name] = make([]int32, 0)
	}
	return nil
}

func (d *MockInfraDriver) DeletePipelineResources(ctx context.Context, pipeline string) error {
	if _, ok := d.Rcs[pipeline]; ok {
		delete(d.Rcs, pipeline)
		d.IncCall(pipeline, MockInfraOp_DELETE)
	}
	return nil
}

func (d *MockInfraDriver) ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationControllerList, error) {
	d.IncCall(pi.Pipeline.Name, MockInfraOp_READ)
	if rc, ok := d.Rcs[pi.Pipeline.Name]; !ok {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{},
		}, errors.New("rc for pipeline not found")
	} else {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{rc},
		}, nil
	}
}

func (d *MockInfraDriver) UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error {
	rc := old.DeepCopy()
	if update(rc) {
		name := rc.ObjectMeta.Labels[pipelineNameLabel]
		d.ScaleHistory[name] = append(d.ScaleHistory[name], *rc.Spec.Replicas)
		d.WriteRC(rc)
	}
	return nil
}

func (d *MockInfraDriver) ListReplicationControllers(ctx context.Context) (*v1.ReplicationControllerList, error) {
	items := make([]v1.ReplicationController, 0)
	for _, rc := range d.Rcs {
		items = append(items, rc)
	}
	return &v1.ReplicationControllerList{
		Items: items,
	}, nil
}

// TODO(acohen4): complete
func (d *MockInfraDriver) WatchPipelinePods(ctx context.Context) (<-chan watch.Event, func(), error) {
	ch := make(chan watch.Event)
	return ch, func() {}, nil
}

////////////////////////////////////
// -------- Mock Helpers -------- //
////////////////////////////////////

func (d *MockInfraDriver) ResetRCs() {
	d.Rcs = make(map[string]v1.ReplicationController)
}

func (d *MockInfraDriver) MakeRC(pi *pps.PipelineInfo) *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name: ppsutil.PipelineRcName(pi.Pipeline.Name, pi.Version),
			Annotations: map[string]string{
				pipelineVersionAnnotation:    strconv.FormatUint(pi.Version, 10),
				pipelineSpecCommitAnnotation: pi.SpecCommit.ID,
				hashedAuthTokenAnnotation:    hashAuthToken(pi.AuthToken),
				pachVersionAnnotation:        version.PrettyVersion(),
			},
			Labels: pipelineLabels(pi.Pipeline.Name, pi.Version),
		},
	}
}

func (d *MockInfraDriver) WriteRC(rc *v1.ReplicationController) {
	name := rc.ObjectMeta.Labels[pipelineNameLabel]
	d.Rcs[name] = *rc
}

func (d *MockInfraDriver) IncCall(name string, op MockInfraOp) {
	if ops, ok := d.Calls[name]; ok {
		ops[op] += 1
	} else {
		d.Calls[name] = map[MockInfraOp]int{op: 1}
	}
}
