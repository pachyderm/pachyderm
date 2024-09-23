package pjs

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pachyderm/pachyderm/v2/src/pjs"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

type Doer struct {
	Client           pjs.APIClient
	InputTranslator  func(context.Context, *anypb.Any) ([]string, error)
	OutputTranslator func(context.Context, []string) (*anypb.Any, error)
	Program          string // fileset ID of job program
}

func (pd Doer) Do(ctx context.Context, inputChan chan *anypb.Any, cb task.CollectFunc) (retErr error) {
	var (
		i  int64
		wg sync.WaitGroup
	)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer wg.Wait()

	for {
		select {
		case input, ok := <-inputChan:
			if !ok { // i.e., input channel is closed
				return
			}
			inputs, err := pd.translateInput(ctx, input)
			if err != nil {
				errors.JoinInto(&retErr, errors.Wrap(err, "translating input"))
				return
			}
			resp, err := pd.Client.CreateJob(ctx, &pjs.CreateJobRequest{
				Program: pd.Program,
				Input:   inputs,
			})
			if err != nil {
				errors.JoinInto(&retErr, errors.Wrap(err, "creating job"))
				return
			}
			wg.Add(1)
			go func(index int64, job *pjs.Job) {
				defer wg.Done()
				if err := backoff.RetryUntilCancel(ctx, func() error {
					resp, err := pd.Client.InspectJob(ctx, &pjs.InspectJobRequest{Job: job})
					if err != nil {
						return errors.Wrap(err, "inspecting job")
					}
					switch result := resp.Details.JobInfo.Result.(type) {
					case *pjs.JobInfo_Success_:
						output, err := pd.translateOutput(ctx, result.Success.Output)
						if err != nil {
							errors.JoinInto(&retErr, err)
							if err := cb(index, nil, err); err != nil {
								return errors.Wrap(err, "calling error callback")
							}
							return errors.Wrap(err, "translating output to proto")
						}
						if err := cb(index, output, nil); err != nil {
							return errors.Wrap(err, "processing output")
						}
						return nil
					case *pjs.JobInfo_Error:
						// FIXME: retry logic here
						return errors.Errorf("don’t know how to handle job error")
					default:
						return errors.Errorf("unexpected job result %T", result)
					}
				}, backoff.NewExponentialBackOff(),
					func(err error, d time.Duration) error {
						log.Info(ctx, "retrying InspectJob", zap.Error(err), zap.Duration("backoff", d))
						return nil
					}); err != nil {
					panic(err)
				}
			}(i, resp.Id)
			i++
		case <-ctx.Done():
			errors.JoinInto(&retErr, ctx.Err())
			return
		}
	}
}

func (pd Doer) translateInput(ctx context.Context, input *anypb.Any) ([]string, error) {
	// FIXME: do something clever if translator missing
	return pd.InputTranslator(ctx, input)
}

func (pd Doer) translateOutput(ctx context.Context, output []string) (*anypb.Any, error) {
	// FIXME: do something clever if translator missing
	return pd.OutputTranslator(ctx, output)
}
