package pjs

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func CodeSpecID(codeSpec *anypb.Any) Sum {
	data, err := proto.Marshal(codeSpec)
	if err != nil {
		panic(err)
	}
	return pachhash.Sum(data)
}

// Do does work through PJS.
func Do[In, Out proto.Message](ctx context.Context, s pjs.APIClient, in *pjs.CreateJobRequest, fn func(*pjs.QueueElement) (*pjs.QueueElement, error)) (*pjs.QueueElement, error) {
	j, err := s.CreateJob(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := ProcessQueue(ctx, s, in.Spec, fn); err != nil {
		return nil, err
	}
	jobInfo, err := Await(ctx, s, JobID(j.Id.Id))
	if err != nil {
		return nil, err
	}
	// TODO: check if failed
	return jobInfo.GetOutput(), nil
}

func ProcessQueue(ctx context.Context, s pjs.APIClient, codeSpec *anypb.Any, fn func(*pjs.QueueElement) (*pjs.QueueElement, error)) error {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	queueID := CodeSpecID(codeSpec)
	pqc, err := s.ProcessQueue(ctx)
	if err != nil {
		return err
	}
	if err := pqc.Send(&pjs.ProcessQueueRequest{
		Queue: &pjs.Queue{Id: queueID[:]},
	}); err != nil {
		return err
	}
	for {
		msg, err := pqc.Recv()
		if err != nil {
			return err
		}
		out, err := fn(msg.Input)
		if err != nil {
			pqc.Send(&pjs.ProcessQueueRequest{
				Result: &pjs.ProcessQueueRequest_Failed{
					Failed: true,
				},
			})
			return err
		} else {
			pqc.Send(&pjs.ProcessQueueRequest{
				Result: &pjs.ProcessQueueRequest_Output{
					Output: out,
				},
			})
		}
	}
}

// Await blocks until a Job enters the DONE state
func Await(ctx context.Context, s pjs.APIClient, jid JobID) (*pjs.JobInfo, error) {
	for {
		res, err := s.InspectJob(ctx, &pjs.InspectJobRequest{
			Job: &pjs.Job{Id: int64(jid)},
		})
		if err != nil {
			return nil, err
		}
		if res.Details.JobInfo.State == pjs.JobState_DONE {
			return res.Details.JobInfo, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func NewTestClient(t testing.TB) pjs.APIClient {
	srv := NewServer(Env{})
	gc := grpcutil.NewTestClient(t, func(s *grpc.Server) {
		pjs.RegisterAPIServer(s, srv)
	})
	return pjs.NewAPIClient(gc)
}
