package server

import (
	"path"
	"strings"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/taskapi"

	etcd "go.etcd.io/etcd/client/v3"
)

type apiServer struct {
	ppsPrefix, pfsPrefix string
	etcdClient           *etcd.Client
}

func NewAPIServer(base, pps, pfs string, etcdClient *etcd.Client) taskapi.APIServer {
	return &apiServer{
		ppsPrefix:  path.Join(base, pps),
		pfsPrefix:  path.Join(base, pfs),
		etcdClient: etcdClient,
	}
}

func translateTaskState(state task.State) taskapi.State {
	switch state {
	case task.State_RUNNING:
		return taskapi.State_RUNNING
	case task.State_SUCCESS:
		return taskapi.State_SUCCESS
	case task.State_FAILURE:
		return taskapi.State_FAILURE
	}
	return taskapi.State_STATE_UNKNOWN
}

func (a *apiServer) ListTask(req *taskapi.ListTaskRequest, server taskapi.API_ListTaskServer) (retError error) {
	ctx := server.Context()
	if req.Namespace == "" {
		return errors.New("Must include a task namespace")
	}
	prefix := a.ppsPrefix
	if strings.HasPrefix(req.Namespace, "storage") {
		prefix = a.pfsPrefix
	}
	etcdCols := task.NewNamespaceEtcd(a.etcdClient, prefix, req.Namespace)
	var taskData task.Task
	etcdCols.TaskCol.ReadOnly(ctx).List(&taskData, col.DefaultOptions(), func(key string) error {
		state := translateTaskState(taskData.State)
		if state == taskapi.State_RUNNING {
			var claim task.Claim
			if etcdCols.ClaimCol.ReadOnly(ctx).Get(key, &claim) == nil {
				state = taskapi.State_CLAIMED
			}
		}
		taskInfo := taskapi.TaskInfo{
			ID:        taskData.ID,
			FullKey:   path.Join(req.Namespace, key),
			State:     state,
			Reason:    taskData.Reason,
			InputType: taskData.Input.TypeUrl,
		}
		return server.Send(&taskInfo)
	})
	return nil
}
