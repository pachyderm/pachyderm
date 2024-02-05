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
	shardNumFiles = 10000
)

type createDatumStream struct {
	ctx               context.Context
	c                 pfs.APIClient
	taskDoer          task.Doer
	input             *pps.Input
	requestDatumsChan chan bool
	fsidChan          chan string
	errChan           chan error
	doneChan          chan bool
}

func (cds *createDatumStream) create(input *pps.Input, fsidCh chan string) {
	switch {
	case input.Pfs != nil:
		cds.createPFS(input.Pfs, fsidCh)
	case input.Union != nil:
		cds.errChan <- errors.New("union input type unimplemented")
	case input.Cross != nil:
		cds.errChan <- errors.New("cross input type unimplemented")
	case input.Join != nil:
		cds.errChan <- errors.New("join input type unimplemented")
	case input.Group != nil:
		cds.errChan <- errors.New("group input type unimplemented")
	case input.Cron != nil:
		cds.errChan <- errors.New("can't create datums for cron input type")
	default:
		cds.errChan <- errors.Errorf("unrecognized input type: %v", input)
	}
}

func (cds *createDatumStream) createPFS(input *pps.PFSInput, fsidChan chan string) {
	defer func() {
		close(cds.doneChan)
		close(fsidChan)
	}()
	authToken := getAuthToken(cds.ctx)
	if err := client.WithRenewer(cds.ctx, cds.c, func(ctx context.Context, renewer *renew.StringSet) error {
		fileSetID, err := client.GetFileSet(ctx, cds.c, input.Project, input.Repo, input.Branch, input.Commit)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return err
		}
		shards, err := client.ShardFileSetWithConfig(ctx, cds.c, fileSetID, shardNumFiles, 0)
		if err != nil {
			return err
		}
		for i, shard := range shards {
			// Block until the client requests more datums
			<-cds.requestDatumsChan
			input, err := serializePFSTask(&PFSTask{
				Input:     input,
				PathRange: shard,
				BaseIndex: createBaseIndex(int64(i)),
				AuthToken: authToken,
			})
			if err != nil {
				return err
			}
			output, err := task.DoOne(ctx, cds.taskDoer, input)
			if err != nil {
				return err
			}
			result, err := deserializePFSTaskResult(output)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return err
			}
			cds.fsidChan <- result.FileSetId
		}
		return nil
	}); err != nil {
		cds.errChan <- err
	}
}
