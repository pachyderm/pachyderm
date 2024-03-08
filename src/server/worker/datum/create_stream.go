package datum

import (
	"context"
	reflect "reflect"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"
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
		streamingCreateUnion(ctx, c, taskDoer, input.Union, fsidChan, errChan, requestDatumsChan)
	case input.Cross != nil:
		streamingCreateCross(ctx, c, taskDoer, input.Cross, fsidChan, errChan, requestDatumsChan)
	case input.Join != nil:
		errChan <- errors.New("join input type unimplemented")
	case input.Group != nil:
		errChan <- errors.New("group input type unimplemented")
	case input.Cron != nil:
		select {
		case <-ctx.Done():
		case errChan <- errors.New("can't create datums for cron input type"):
		}
	default:
		select {
		case <-ctx.Done():
		case errChan <- errors.Errorf("unrecognized input type: %v", input):
		}
	}
}

func streamingCreatePFS(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	input *pps.PFSInput,
	fsidChan chan<- string,
	errChan chan<- error,
	requestDatumsChan <-chan struct{},
) {
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
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreatePFS"):
		}
	}
}

func streamingCreateUnion(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	inputs []*pps.Input,
	fsidChan chan<- string,
	errChan chan<- error,
	requestDatumsChan <-chan struct{},
) {
	defer close(fsidChan)
	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
		childrenChans := initChildrenChans(len(inputs))
		go streamingCreateInputs(ctx, c, taskDoer, renewer, inputs, errChan, requestDatumsChan, childrenChans)
		err := consumeChildrenChans(ctx, childrenChans, func(_ int, fsid string) error {
			fsidChan <- fsid
			return nil
		})
		return errors.Wrap(err, "consuming children channels")
	}); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateUnion"):
		}
	}
}

func initChildrenChans(num int) []chan string {
	childrenFsidChans := make([]chan string, num)
	for i := 0; i < num; i++ { // TODO: go 1.22
		childrenFsidChans[i] = make(chan string)
	}
	return childrenFsidChans
}

func consumeChildrenChans(ctx context.Context, childrenChans []chan string, cb func(int, string) error) error {
	cases := make([]reflect.SelectCase, len(childrenChans)+1)
	for i, childChan := range childrenChans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(childChan)}
	}
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
	finished := make(map[int]bool)
	for len(finished) < len(cases)-1 {
		i, value, ok := reflect.Select(cases)
		if i == len(cases)-1 {
			return errors.Wrap(context.Cause(ctx), "consumeChildrenChans")
		}
		if ok {
			fsid := value.Interface().(string)
			if err := cb(i, fsid); err != nil {
				return errors.Wrap(err, "consumeChildrenChans")
			}
		} else {
			finished[i] = true
		}
	}
	return nil
}

func streamingCreateInputs(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	renewer *renew.StringSet,
	inputs []*pps.Input,
	errChan chan<- error,
	requestDatumsChan <-chan struct{},
	childrenChans []chan string,
) {
	eg, egCtx := errgroup.WithContext(ctx)
	for i, input := range inputs {
		eg.Go(func() error {
			fsidChan := make(chan string)
			go streamingCreate(egCtx, c, taskDoer, input, fsidChan, errChan, requestDatumsChan)
			for fsid := range fsidChan {
				if err := renewer.Add(ctx, fsid); err != nil {
					return errors.Wrap(err, "renew file set")
				}
				childrenChans[i] <- fsid
			}
			close(childrenChans[i])
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateInputs"):
		}
	}
}

func streamingCreateCross(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	inputs []*pps.Input,
	fsidChan chan<- string,
	errChan chan<- error,
	requestDatumsChan <-chan struct{},
) {
	defer close(fsidChan)
	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
		childrenChans := initChildrenChans(len(inputs))
		go streamingCreateInputs(ctx, c, taskDoer, renewer, inputs, errChan, requestDatumsChan, childrenChans)
		inputsShards := make([][]string, len(inputs))
		err := consumeChildrenChans(childrenChans, func(i int, fsid string) error {
			inputsShards[i] = append(inputsShards[i], fsid)
			taskInputs, err := crossShards(ctx, c, inputsShards, i)
			if err != nil {
				return errors.Wrap(err, "cross shards")
			}
			if taskInputs != nil {
				if err := task.DoBatch(ctx, taskDoer, taskInputs, func(i int64, output *anypb.Any, err error) error {
					if err != nil {
						return err
					}
					result, err := deserializeCrossTaskResult(output)
					if err != nil {
						return errors.Wrap(err, "deserialize cross task result")
					}
					if err := renewer.Add(ctx, result.FileSetId); err != nil {
						return errors.Wrap(err, "renew result file set")
					}
					select {
					case <-ctx.Done():
						return errors.Wrap(context.Cause(ctx), "send file set id to iterator")
					case fsidChan <- result.FileSetId:
					}
					return nil
				}); err != nil {
					return errors.Wrap(err, "task do batch")
				}
			}
			return nil
		})
		return errors.Wrap(err, "consuming children channels")
	}); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateCross"):
		}
	}
}

// crossShards returns a slice of task inputs, each of which represents a cross task. Each
// cross task is a permutation of shards from each input. For the input that just received
// a shard (addedInput), cross that with shards from all other inputs.
func crossShards(ctx context.Context, c pfs.APIClient, inputsShards [][]string, addedInput int) ([]*anypb.Any, error) {
	// If any input has no shards, we can't cross it with the other inputs
	for _, inputShards := range inputsShards {
		if len(inputShards) == 0 {
			return nil, nil
		}
	}
	baseFileSetShards, err := client.ShardFileSetWithConfig(ctx, c, inputsShards[addedInput][len(inputsShards[addedInput])-1], ShardNumFiles, 0)
	if err != nil {
		return nil, errors.Wrap(err, "shard file set")
	}
	shardsCrosses := generateShardPermutations(inputsShards, addedInput)
	var taskInputs []*anypb.Any
	for _, shardsCross := range shardsCrosses {
		for i, shard := range baseFileSetShards {
			input, err := serializeCrossTask(&CrossTask{
				FileSetIds:           shardsCross,
				BaseFileSetIndex:     int64(addedInput),
				BaseFileSetPathRange: shard,
				BaseIndex:            createBaseIndex(int64(i)),
				AuthToken:            getAuthToken(ctx),
			})
			if err != nil {
				return nil, errors.Wrap(err, "serialize cross task")
			}
			taskInputs = append(taskInputs, input)
		}
	}
	return taskInputs, nil
}

// Generates all permutations of shards from each input. From the addedInput input, only the
// most recently added shard is used. Ex:
//
// inputsShards = [[a1, a2], [b1, b2], [c1, c2]]
//
// addedInput = 1
//
// output = [[a1, b2, c1], [a1, b2, c2], [a2, b2, c1], [a2, b2, c2]]
//
// The b2 shard associated with addedInput 1 is used in all permutations.
func shardPermute(input [][]string, addedInput int, index int, result []string, output *[][]string) {
	if index == len(input) {
		temp := make([]string, len(result))
		copy(temp, result)
		*output = append(*output, temp)
		return
	}
	if index == addedInput {
		result[index] = input[index][len(input[index])-1]
		shardPermute(input, addedInput, index+1, result, output)
		return
	}
	for i := 0; i < len(input[index]); i++ {
		result[index] = input[index][i]
		shardPermute(input, addedInput, index+1, result, output)
	}
}

// Used for crossing shards. inputsShards[i] refers to a slice of shards for the ith input.
// addedInput is the index of the input that most recently received a shard. This function
// returns a slice of slices, each of which represents a permutation of shards from each input.
func generateShardPermutations(inputsShards [][]string, addedInput int) [][]string {
	output := make([][]string, 0)
	result := make([]string, len(inputsShards))
	shardPermute(inputsShards, addedInput, 0, result, &output)
	return output
}
