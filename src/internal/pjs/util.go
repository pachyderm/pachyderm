package pjs

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/grpc"
)

type ClientOptions func(env *Env)

func NewTestClient(t testing.TB, db *pachsql.DB, storage *storage.Server, opts ...ClientOptions) pjs.APIClient {
	env := Env{
		DB:               db,
		GetPermissionser: &testPermitter{mode: permitterAllow},
		GetAuthToken:     func(ctx context.Context) (string, error) { return tu.RootToken, nil },
		Storage:          storage,
	}
	for _, opt := range opts {
		opt(&env)
	}
	srv := NewAPIServer(env)
	gc := grpcutil.NewTestClient(t, func(s *grpc.Server) {
		pjs.RegisterAPIServer(s, srv)
	})
	return pjs.NewAPIClient(gc)
}

type permitterEnum int

const (
	permitterAllow permitterEnum = iota
	permitterDeny
)

type testPermitter struct {
	override *auth.GetPermissionsResponse
	mode     permitterEnum
}

func (p *testPermitter) GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error) {
	switch {
	case p.override != nil:
		return p.override, nil
	case p.mode == permitterAllow:
		return &auth.GetPermissionsResponse{
			Permissions: []auth.Permission{auth.Permission_JOB_SKIP_CTX},
			Roles:       []string{auth.ClusterAdminRole},
		}, nil
	case p.mode == permitterDeny:
		return &auth.GetPermissionsResponse{}, nil
	default:
		return nil, nil
	}
}

func ToJobInfo(ctx context.Context, tx *pachsql.Tx, storage *fileset.Storage, job pjsdb.Job) (*pjs.JobInfo, error) {
	jobInfo := &pjs.JobInfo{
		Job: &pjs.Job{
			Id: int64(job.ID),
		},
		ParentJob: &pjs.Job{
			Id: int64(job.Parent),
		},
	}
	programHandle, err := storage.GetPinHandleTx(ctx, tx, job.Program, defaultTTL)
	if err != nil {
		return nil, err
	}
	jobInfo.Program = programHandle.HexString()
	for _, input := range job.Inputs {
		handle, err := storage.GetPinHandleTx(ctx, tx, input, defaultTTL)
		if err != nil {
			return nil, err
		}
		jobInfo.Input = append(jobInfo.Input, handle.HexString())
	}
	switch {
	case job.Done != time.Time{}:
		jobInfo.State = pjs.JobState_DONE
		jobInfo.Result = &pjs.JobInfo_Error{
			Error: pjsdb.EnumStringToErrorCode[job.Error],
		}
		if len(job.Outputs) != 0 {
			jobInfoSuccess := pjs.JobInfo_Success{}
			for _, output := range job.Outputs {
				handle, err := storage.GetPinHandleTx(ctx, tx, output, defaultTTL)
				if err != nil {
					return nil, err
				}
				jobInfoSuccess.Output = append(jobInfoSuccess.Output, handle.HexString())
			}
			jobInfo.Result = &pjs.JobInfo_Success_{Success: &jobInfoSuccess}
		}
	case job.Processing != time.Time{}:
		jobInfo.State = pjs.JobState_PROCESSING
	default:
		jobInfo.State = pjs.JobState_QUEUED
	}
	return jobInfo, nil
}

func ToQueueInfo(ctx context.Context, tx *pachsql.Tx, storage *fileset.Storage, queue pjsdb.Queue) (*pjs.QueueInfo, error) {
	handle, err := storage.GetPinHandleTx(ctx, tx, queue.Program, defaultTTL)
	if err != nil {
		return nil, err
	}
	return &pjs.QueueInfo{
		Queue: &pjs.Queue{
			Id: queue.ID,
		},
		Program: handle.HexString(),
	}, nil
}
