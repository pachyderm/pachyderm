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
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
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
		streamingCreateJoin(ctx, c, taskDoer, input.Join, fsidChan, errChan, requestDatumsChan)
	case input.Group != nil:
		streamingCreateGroup(ctx, c, taskDoer, input.Group, fsidChan, errChan)
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
				AuthToken: getAuthToken(ctx),
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
				return errors.Wrap(context.Cause(ctx), "send file set id up")
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
		childrenFsidChans := initChildrenFsidChans(len(inputs))
		go streamingCreateInputs(ctx, c, taskDoer, renewer, inputs, childrenFsidChans, errChan, requestDatumsChan)
		if err := consumeChildrenFsidChans(ctx, childrenFsidChans, func(_ int, fsid string) error {
			fsidChan <- fsid
			return nil
		}); err != nil {
			return errors.Wrap(err, "consuming children channels")
		}
		return nil
	}); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateUnion"):
		}
	}
}

// Creates a slice of channels for each child input. Each channel will receive file set id shards
func initChildrenFsidChans(num int) []chan string {
	childrenFsidChans := make([]chan string, num)
	for i := range num {
		childrenFsidChans[i] = make(chan string)
	}
	return childrenFsidChans
}

// Consumes file set id shards from each child input as they arrive. Calls cb with the index of
// the child input. Function returns when all children channels are closed or the context is done.
func consumeChildrenFsidChans(ctx context.Context, childrenFsidChans []chan string, cb func(int, string) error) error {
	cases := make([]reflect.SelectCase, len(childrenFsidChans)+1)
	for i, childChan := range childrenFsidChans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(childChan)}
	}
	cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
	finished := make(map[int]bool)
	for len(finished) < len(cases)-1 {
		i, value, ok := reflect.Select(cases)
		if i == len(cases)-1 {
			return errors.Wrap(context.Cause(ctx), "consumeChildrenFsidChans")
		}
		if ok {
			fsid := value.Interface().(string)
			if err := cb(i, fsid); err != nil {
				return errors.Wrap(err, "consumeChildrenFsidChans")
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
	childrenFsidChans []chan string,
	errChan chan<- error,
	requestDatumsChan <-chan struct{},
) {
	eg, egCtx := errgroup.WithContext(ctx)
	for i, input := range inputs {
		eg.Go(func() error {
			fsidChan := make(chan string)
			go streamingCreate(egCtx, c, taskDoer, input, fsidChan, errChan, requestDatumsChan)
			for fsid := range fsidChan {
				if err := renewer.Add(egCtx, fsid); err != nil {
					return errors.Wrap(err, "renew file set")
				}
				childrenFsidChans[i] <- fsid
			}
			close(childrenFsidChans[i])
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
		childrenFsidChans := initChildrenFsidChans(len(inputs))
		go streamingCreateInputs(ctx, c, taskDoer, renewer, inputs, childrenFsidChans, errChan, requestDatumsChan)
		inputsShards := make([][]string, len(inputs))
		if err := consumeChildrenFsidChans(ctx, childrenFsidChans, func(idx int, fsid string) error {
			inputsShards[idx] = append(inputsShards[idx], fsid)
			taskInputs, err := getTasksForNewShard(ctx, c, inputsShards, idx, func(ctx context.Context, fileSetIDs []string, baseShardIndex int, baseShard *pfs.PathRange) (*anypb.Any, error) {
				input, err := serializeCrossTask(&CrossTask{
					FileSetIds:           fileSetIDs,
					BaseFileSetIndex:     int64(idx),
					BaseFileSetPathRange: baseShard,
					BaseIndex:            createBaseIndex(int64(baseShardIndex)),
					AuthToken:            getAuthToken(ctx),
				})
				if err != nil {
					return nil, errors.Wrap(err, "serialize cross task")
				}
				return input, nil
			})
			if err != nil {
				return errors.Wrap(err, "cartesian product")
			}
			if taskInputs != nil {
				return task.DoBatch(ctx, taskDoer, taskInputs, func(_ int64, output *anypb.Any, err error) error {
					if err != nil {
						return errors.Wrap(err, "task do batch")
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
						return errors.Wrap(context.Cause(ctx), "send file set id up")
					case fsidChan <- result.FileSetId:
					}
					return nil
				})
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "consuming children channels")
		}
		return nil
	}); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateCross"):
		}
	}
}

// Returns a slice of task inputs. Each task is a permutation of a shard from each input.
// For the input that just received a shard (addedInput), use only the recently added
// shard. For the other inputs, use all shards.
func getTasksForNewShard(
	ctx context.Context,
	c pfs.APIClient,
	inputsShards [][]string,
	addedInput int,
	serializeCb func(context.Context, []string, int, *pfs.PathRange) (*anypb.Any, error),
) ([]*anypb.Any, error) {
	shardPermutations := cartesianProduct(inputsShards, addedInput)
	if shardPermutations == nil {
		return nil, nil
	}
	baseShards, err := client.ShardFileSetWithConfig(ctx, c, inputsShards[addedInput][len(inputsShards[addedInput])-1], ShardNumFiles, 0)
	if err != nil {
		return nil, errors.Wrap(err, "shard file set")
	}
	var taskInputs []*anypb.Any
	for _, shardPermutation := range shardPermutations {
		for i, baseShard := range baseShards {
			input, err := serializeCb(ctx, shardPermutation, i, baseShard)
			if err != nil {
				return nil, errors.Wrap(err, "serialize callback")
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
func shardPermute(inputsShards [][]string, addedInput int, index int, result []string, output *[][]string) {
	if index == len(inputsShards) {
		temp := append([]string(nil), result...)
		*output = append(*output, temp)
		return
	}
	if index == addedInput {
		result[index] = inputsShards[index][len(inputsShards[index])-1]
		shardPermute(inputsShards, addedInput, index+1, result, output)
		return
	}
	for i := 0; i < len(inputsShards[index]); i++ {
		result[index] = inputsShards[index][i]
		shardPermute(inputsShards, addedInput, index+1, result, output)
	}
}

// Used for crossing shards. inputsShards[i] refers to a slice of shards for the ith input.
// addedInput is the index of the input that most recently received a shard. This function
// returns a slice of slices, each of which represents a permutation of shards from each input,
// except for the added input, which only uses the most recently added shard.
func cartesianProduct(inputsShards [][]string, addedInput int) [][]string {
	// If any input has no shards, we can't cross it with the other inputs
	for _, inputShards := range inputsShards {
		if len(inputShards) == 0 {
			return nil
		}
	}
	output := make([][]string, 0)
	result := make([]string, len(inputsShards))
	shardPermute(inputsShards, addedInput, 0, result, &output)
	return output
}

func streamingCreateJoin(
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
		childrenFsidChans := initChildrenFsidChans(len(inputs))
		go streamingCreateInputs(ctx, c, taskDoer, renewer, inputs, childrenFsidChans, errChan, requestDatumsChan)
		inputsShards := make([][]string, len(inputs))
		if err := consumeChildrenFsidChans(ctx, childrenFsidChans, func(idx int, fsid string) error {
			keyFsidChan := make(chan string)
			go streamingCreateKeyFileSetsJoin(ctx, c, taskDoer, renewer, keyFsidChan, errChan, fsid)
			for keyFsid := range keyFsidChan {
				inputsShards[idx] = append(inputsShards[idx], keyFsid)
				if err := streamingMergeKeyFileSetsJoin(ctx, c, taskDoer, renewer, fsidChan, inputsShards, idx); err != nil {
					return errors.Wrap(err, "streamingMergeKeyFileSetsJoin")
				}
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "consuming children channels")
		}
		return nil
	}); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateJoin"):
		}
	}
}

func streamingCreateKeyFileSetsJoin(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	renewer *renew.StringSet,
	fsidChan chan<- string,
	errChan chan<- error,
	fileSetID string,
) {
	defer close(fsidChan)
	if err := func() error {
		shards, err := client.ShardFileSetWithConfig(ctx, c, fileSetID, ShardNumFiles, 0)
		if err != nil {
			return errors.Wrap(err, "shard file set")
		}
		var inputs []*anypb.Any
		for _, shard := range shards {
			input, err := serializeKeyTask(&KeyTask{
				FileSetId: fileSetID,
				PathRange: shard,
				Type:      KeyTask_JOIN,
				AuthToken: getAuthToken(ctx),
			})
			if err != nil {
				return errors.Wrap(err, "serialize key task")
			}
			inputs = append(inputs, input)
		}
		return task.DoBatch(ctx, taskDoer, inputs, func(_ int64, output *anypb.Any, err error) error {
			if err != nil {
				return errors.Wrap(err, "task do batch")
			}
			result, err := deserializeKeyTaskResult(output)
			if err != nil {
				return errors.Wrap(err, "deserialize key task result")
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return errors.Wrap(err, "renew result file set")
			}
			select {
			case <-ctx.Done():
				return errors.Wrap(context.Cause(ctx), "send file set id up")
			case fsidChan <- result.FileSetId:
			}
			return nil
		})
	}(); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateKeyFileSet"):
		}
	}
}

func streamingMergeKeyFileSetsJoin(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	renewer *renew.StringSet,
	fsidChan chan<- string,
	inputsShards [][]string,
	idx int,
) error {
	taskInputs, err := getTasksForNewShard(ctx, c, inputsShards, idx, func(ctx context.Context, fileSetIDs []string, _ int, baseShard *pfs.PathRange) (*anypb.Any, error) {
		input, err := serializeMergeTask(&MergeTask{
			FileSetIds: fileSetIDs,
			PathRange:  baseShard,
			Type:       MergeTask_JOIN,
			AuthToken:  getAuthToken(ctx),
		})
		if err != nil {
			return nil, errors.Wrap(err, "serialize merge task")
		}
		return input, nil
	})
	if err != nil {
		return errors.Wrap(err, "cartesian product")
	}
	if taskInputs != nil {
		return task.DoBatch(ctx, taskDoer, taskInputs, func(_ int64, output *anypb.Any, err error) error {
			if err != nil {
				return errors.Wrap(err, "task do batch")
			}
			result, err := deserializeMergeTaskResult(output)
			if err != nil {
				return errors.Wrap(err, "deserialize merge task result")
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return errors.Wrap(err, "renew result file set")
			}
			select {
			case <-ctx.Done():
				return errors.Wrap(context.Cause(ctx), "send file set id up")
			case fsidChan <- result.FileSetId:
			}
			return nil
		})
	}
	return nil
}

func streamingCreateGroup(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	inputs []*pps.Input,
	fsidChan chan<- string,
	errChan chan<- error,
) {
	defer close(fsidChan)
	if err := client.WithRenewer(ctx, c, func(ctx context.Context, renewer *renew.StringSet) error {
		// Group needs all key file set shards to be present before merging. Use ListDatum
		// implementation until merge step, instead of streaming, as it's faster.
		fileSetIDs, err := createInputs(ctx, c, taskDoer, renewer, inputs)
		if err != nil {
			return err
		}
		keyFileSetIDs, err := createKeyFileSets(ctx, c, taskDoer, renewer, fileSetIDs, KeyTask_GROUP)
		if err != nil {
			return err
		}
		if err := streamingMergeKeyFileSetsGroup(ctx, c, taskDoer, renewer, fsidChan, keyFileSetIDs); err != nil {
			return errors.Wrap(err, "streamingMergeKeyFileSetsGroup")
		}
		return nil
	}); err != nil {
		select {
		case <-ctx.Done():
		case errChan <- errors.Wrap(err, "streamingCreateGroup"):
		}
	}
}

func streamingMergeKeyFileSetsGroup(
	ctx context.Context,
	c pfs.APIClient,
	taskDoer task.Doer,
	renewer *renew.StringSet,
	fsidChan chan<- string,
	composedShards []string,
) error {
	shards, err := common.Shard(ctx, c, composedShards)
	if err != nil {
		return errors.Wrap(err, "shard file set")
	}
	var inputs []*anypb.Any
	for _, shard := range shards {
		input, err := serializeMergeTask(&MergeTask{
			FileSetIds: composedShards,
			PathRange:  shard,
			Type:       MergeTask_GROUP,
			AuthToken:  getAuthToken(ctx),
		})
		if err != nil {
			return errors.Wrap(err, "serialize merge task")
		}
		inputs = append(inputs, input)
	}
	return task.DoBatch(ctx, taskDoer, inputs, func(_ int64, output *anypb.Any, err error) error {
		if err != nil {
			return errors.Wrap(err, "task do batch")
		}
		result, err := deserializeMergeTaskResult(output)
		if err != nil {
			return errors.Wrap(err, "deserialize merge task result")
		}
		if err := renewer.Add(ctx, result.FileSetId); err != nil {
			return errors.Wrap(err, "renew result file set")
		}
		select {
		case <-ctx.Done():
			return errors.Wrap(context.Cause(ctx), "send file set id up")
		case fsidChan <- result.FileSetId:
		}
		return nil
	})
}
