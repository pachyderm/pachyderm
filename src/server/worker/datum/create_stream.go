package datum

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// The target number of files per shard. A smaller value returns datums to
	// the client faster. A larger value reduces the overhead of creating datums.
	ShardNumFiles = 10000
)

// Creates datums from the given input and sends the file set ids over fsidChan
// as they are created. Should be called in a goroutine.
func streamingCreate(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.Input, fsidChan chan<- string, errChan chan<- error, requestDatumsChan <-chan struct{}) {
	switch {
	case input.Pfs != nil:
		streamingCreatePFS(ctx, c, taskDoer, input.Pfs, fsidChan, errChan, requestDatumsChan)
	case input.Union != nil:
		errChan <- errors.New("union input type unimplemented")
	case input.Cross != nil:
		errChan <- errors.New("cross input type unimplemented")
	case input.Join != nil:
		errChan <- errors.New("join input type unimplemented")
	case input.Group != nil:
		errChan <- errors.New("group input type unimplemented")
	case input.Cron != nil:
		errChan <- errors.New("can't create datums for cron input type")
	default:
		errChan <- errors.Errorf("unrecognized input type: %v", input)
	}
}

func streamingCreatePFS(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.PFSInput, fsidChan chan<- string, errChan chan<- error, requestDatumsChan <-chan struct{}) {
	defer close(fsidChan)
	authToken := getAuthToken(ctx)
	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
		fileSetID, err := client.GetFileSet(ctx, c, input.Project, input.Repo, input.Branch, input.Commit)
		if err != nil {
			return errors.Wrap(err, "get file set")
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return errors.Wrap(err, "renew file set")
		}
		shards, err := client.ShardFileSetWithConfig(ctx, c, fileSetID, ShardNumFiles, 0)
		if err != nil {
			return errors.Wrap(err, "shard file set")
		}
		for i, shard := range shards {
			// Block until the client requests more datums
			<-requestDatumsChan
			input, err := serializePFSTask(&PFSTask{
				Input:     input,
				PathRange: shard,
				BaseIndex: createBaseIndex(int64(i)),
				AuthToken: authToken,
			})
			if err != nil {
				return errors.Wrap(err, "serialize pfs task")
			}
			output, err := task.DoOne(ctx, taskDoer, input)
			if err != nil {
				return errors.Wrap(err, "do task")
			}
			result, err := deserializePFSTaskResult(output)
			if err != nil {
				return errors.Wrap(err, "deserialize pfs task result")
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return errors.Wrap(err, "renew result file set")
			}
			select {
			case <-ctx.Done():
				return errors.Wrap(context.Cause(ctx), "send file set id to iterator")
			case fsidChan <- result.FileSetId:
			}

		}
		return nil
	}); err != nil {
		errChan <- errors.Wrap(err, "streamingCreatePFS")
	}
}
