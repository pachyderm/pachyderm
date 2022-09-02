package server

import (
	"context"
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
	DeletePipelineResources(ctx context.Context, pipeline *pps.Pipeline) error
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
	rcs          map[pipelineKey]v1.ReplicationController
	calls        map[pipelineKey]map[mockInfraOp]int
	scaleHistory map[pipelineKey][]int32
}

func newMockInfraDriver() *mockInfraDriver {
	d := &mockInfraDriver{
		rcs:          make(map[pipelineKey]v1.ReplicationController),
		calls:        make(map[pipelineKey]map[mockInfraOp]int),
		scaleHistory: make(map[pipelineKey][]int32),
	}
	return d
}

func (d *mockInfraDriver) CreatePipelineResources(ctx context.Context, pi *pps.PipelineInfo) error {
	d.rcs[toKey(pi.Pipeline)] = *d.makeRC(pi)
	d.incCall(toKey(pi.Pipeline), mockInfraOp_CREATE)
	if _, ok := d.scaleHistory[toKey(pi.Pipeline)]; !ok {
		d.scaleHistory[toKey(pi.Pipeline)] = make([]int32, 0)
	}
	return nil
}

func (d *mockInfraDriver) DeletePipelineResources(ctx context.Context, pipeline *pps.Pipeline) error {
	if _, ok := d.rcs[toKey(pipeline)]; ok {
		delete(d.rcs, toKey(pipeline))
		d.incCall(toKey(pipeline), mockInfraOp_DELETE)
	}
	return nil
}

func (d *mockInfraDriver) ReadReplicationController(ctx context.Context, pi *pps.PipelineInfo) (*v1.ReplicationControllerList, error) {
	d.incCall(toKey(pi.Pipeline), mockInfraOp_READ)
	if rc, ok := d.rcs[toKey(pi.Pipeline)]; !ok {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{},
		}, errors.New("rc for pipeline not found")
	} else {
		return &v1.ReplicationControllerList{
			Items: []v1.ReplicationController{rc},
		}, nil
	}
}

func (d *mockInfraDriver) UpdateReplicationController(ctx context.Context, old *v1.ReplicationController, update func(rc *v1.ReplicationController) bool) error {
	rc := old.DeepCopy()
	if update(rc) {
		name := rc.ObjectMeta.Labels[pipelineNameLabel]
		project := rc.ObjectMeta.Labels[pipelineProjectLabel]
		pipeline := &pps.Pipeline{
			Project: &pfs.Project{Name: project},
			Name:    name,
		}
		d.scaleHistory[toKey(pipeline)] = append(d.scaleHistory[toKey(pipeline)], *rc.Spec.Replicas)
		d.writeRC(rc)
	}
	return nil
}

func (d *mockInfraDriver) ListReplicationControllers(ctx context.Context) (*v1.ReplicationControllerList, error) {
	items := make([]v1.ReplicationController, 0)
	for _, rc := range d.rcs {
		items = append(items, rc)
	}
	return &v1.ReplicationControllerList{
		Items: items,
	}, nil
}

// TODO(acohen4): complete
func (d *mockInfraDriver) WatchPipelinePods(ctx context.Context) (<-chan watch.Event, func(), error) {
	ch := make(chan watch.Event)
	return ch, func() {}, nil
}

////////////////////////////////////
// -------- Mock Helpers -------- //
////////////////////////////////////

func (d *mockInfraDriver) resetRCs() {
	d.rcs = make(map[pipelineKey]v1.ReplicationController)
}

func (d *mockInfraDriver) makeRC(pi *pps.PipelineInfo) *v1.ReplicationController {
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

func (d *mockInfraDriver) writeRC(rc *v1.ReplicationController) {
	d.rcs[toKey(&pps.Pipeline{
		Project: &pfs.Project{Name: rc.ObjectMeta.Labels[pipelineProjectLabel]},
		Name:    rc.ObjectMeta.Labels[pipelineNameLabel],
	})] = *rc
}

func (d *mockInfraDriver) incCall(key pipelineKey, op mockInfraOp) {
	if ops, ok := d.calls[key]; ok {
		ops[op] += 1
	} else {
		d.calls[key] = map[mockInfraOp]int{op: 1}
	}
}
