package datum

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"golang.org/x/sync/errgroup"
)

func Create(pachClient *client.APIClient, taskDoer task.Doer, input *pps.Input) (string, error) {
	// TODO: Prefix index.
	switch {
	case input.Pfs != nil:
		return createPFS(pachClient, taskDoer, input.Pfs)
	case input.Union != nil:
		return createUnion(pachClient, taskDoer, input.Union)
	case input.Cross != nil:
		return createCross(pachClient, taskDoer, input.Cross)
	case input.Join != nil:
		return createJoin(pachClient, taskDoer, input.Join)
	case input.Group != nil:
		return createGroup(pachClient, taskDoer, input.Group)
	case input.Cron != nil:
		return createCron(pachClient, taskDoer, input.Cron)
	default:
		return "", errors.Errorf("unrecognized input type: %v", input)
	}
}

func createPFS(pachClient *client.APIClient, taskDoer task.Doer, input *pps.PFSInput) (string, error) {
	if input.Commit == "" {
		// this can happen if a pipeline with multiple inputs has been triggered
		// before all commits have inputs
		return CreateEmptyFileSet(pachClient)
	}
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		fileSetID, err := pachClient.GetFileSet(input.Repo, input.Branch, input.Commit)
		if err != nil {
			return err
		}
		if err := renewer.Add(ctx, fileSetID); err != nil {
			return err
		}
		shards, err := pachClient.ShardFileSet(fileSetID)
		if err != nil {
			return err
		}
		var inputs []*types.Any
		for _, shard := range shards {
			input, err := serializePFSTask(&PFSTask{
				Input:     input,
				PathRange: shard,
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
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
			resultFileSetIDs[i] = result.FileSetId
			return nil
		}); err != nil {
			return err
		}
		resp, err := pachClient.PfsAPIClient.ComposeFileSet(
			ctx,
			&pfs.ComposeFileSetRequest{
				FileSetIds: resultFileSetIDs,
				TtlSeconds: int64(common.TTL),
				Compact:    true,
			},
		)
		if err != nil {
			return err
		}
		outputFileSetID = resp.FileSetId
		return nil
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createUnion(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		var fileSetIDs []string
		for _, input := range inputs {
			fileSetID, err := Create(pachClient, taskDoer, input)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, fileSetID); err != nil {
				return err
			}
			fileSetIDs = append(fileSetIDs, fileSetID)
		}
		resp, err := pachClient.PfsAPIClient.ComposeFileSet(
			ctx,
			&pfs.ComposeFileSetRequest{
				FileSetIds: fileSetIDs,
				TtlSeconds: int64(common.TTL),
				Compact:    true,
			},
		)
		if err != nil {
			return err
		}
		outputFileSetID = resp.FileSetId
		return nil
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createCross(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		var fileSetIDs []string
		for _, input := range inputs {
			fileSetID, err := Create(pachClient, taskDoer, input)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, fileSetID); err != nil {
				return err
			}
			fileSetIDs = append(fileSetIDs, fileSetID)
		}
		var maxIdx int
		var baseFileSetID string
		var baseFileSetShards []*pfs.PathRange
		for i, fileSetID := range fileSetIDs {
			shards, err := pachClient.ShardFileSet(fileSetID)
			if err != nil {
				return err
			}
			if len(shards) > len(baseFileSetShards) {
				maxIdx = i
				baseFileSetID = fileSetID
				baseFileSetShards = shards
			}
		}
		fileSetIDs = append(fileSetIDs[:maxIdx], fileSetIDs[maxIdx+1:]...)
		var inputs []*types.Any
		for _, shard := range baseFileSetShards {
			input, err := serializeCrossTask(&CrossTask{
				BaseFileSetId:        baseFileSetID,
				BaseFileSetPathRange: shard,
				FileSetIds:           fileSetIDs,
			})
			if err != nil {
				return err
			}
			inputs = append(inputs, input)
		}
		resultFileSetIDs := make([]string, len(inputs))
		if err := task.DoBatch(ctx, taskDoer, inputs, func(i int64, output *types.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializeCrossTaskResult(output)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, result.FileSetId); err != nil {
				return err
			}
			resultFileSetIDs[i] = result.FileSetId
			return nil
		}); err != nil {
			return err
		}
		resp, err := pachClient.PfsAPIClient.ComposeFileSet(
			ctx,
			&pfs.ComposeFileSetRequest{
				FileSetIds: resultFileSetIDs,
				TtlSeconds: int64(common.TTL),
				Compact:    true,
			},
		)
		if err != nil {
			return err
		}
		outputFileSetID = resp.FileSetId
		return nil
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

// TODO: Convert to distributed.
func createJoin(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		var iterators []Iterator
		for _, input := range inputs {
			fileSetID, err := Create(pachClient, taskDoer, input)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, fileSetID); err != nil {
				return err
			}
			iterators = append(iterators, NewFileSetIterator(pachClient, fileSetID, nil))
		}
		filesets, err := computeDatumKeyFilesets(pachClient, renewer, iterators, true)
		if err != nil {
			return err
		}
		var dits []Iterator
		for _, fs := range filesets {
			dits = append(dits, newFileSetMultiIterator(pachClient, fs))
		}
		outputFileSetID, err = WithCreateFileSet(pachClient, "pachyderm-datums-join", func(s *Set) error {
			return mergeByKey(dits, existingMetaHash, func(metas []*Meta) error {
				var crossInputs [][]*common.Input
				for _, m := range metas {
					crossInputs = append(crossInputs, m.Inputs)
				}
				err := newCrossListIterator(crossInputs).Iterate(func(meta *Meta) error {
					if len(meta.Inputs) == len(iterators) {
						// all inputs represented, include all inputs
						return s.UploadMeta(meta)
					}
					var filtered []*common.Input
					for _, in := range meta.Inputs {
						if in.OuterJoin {
							filtered = append(filtered, in)
						}
					}
					if len(filtered) > 0 {
						return s.UploadMeta(&Meta{Inputs: filtered})
					}
					return nil
				})
				return errors.EnsureStack(err)
			})
		})
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func computeDatumKeyFilesets(pachClient *client.APIClient, renewer *renew.StringSet, iterators []Iterator, isJoin bool) ([]string, error) {
	eg, ctx := errgroup.WithContext(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)
	filesets := make([]string, len(iterators))
	for i, di := range iterators {
		i := i
		di := di
		keyHasher := pfs.NewHash()
		marshaller := new(jsonpb.Marshaler)
		eg.Go(func() error {
			resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
				err := di.Iterate(func(meta *Meta) error {
					for _, input := range meta.Inputs {
						// hash input keys to ensure consistently-shaped filepaths
						keyHasher.Reset()
						rawKey := input.GroupBy
						if isJoin {
							rawKey = input.JoinOn
						}
						keyHasher.Write([]byte(rawKey))
						key := hex.EncodeToString(keyHasher.Sum(nil))
						out, err := marshaller.MarshalToString(input)
						if err != nil {
							return errors.Wrap(err, "marshalling input for key aggregation")
						}
						if err := mf.PutFile(key, strings.NewReader(out), client.WithAppendPutFile()); err != nil {
							return errors.EnsureStack(err)
						}
					}
					return nil
				})
				return errors.EnsureStack(err)
			})
			if err != nil {
				return err
			}
			renewer.Add(ctx, resp.FileSetId)
			filesets[i] = resp.FileSetId
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return filesets, nil
}

func existingMetaHash(meta *Meta) string {
	return meta.Hash
}

// TODO: Convert to distributed.
func createGroup(pachClient *client.APIClient, taskDoer task.Doer, inputs []*pps.Input) (string, error) {
	var outputFileSetID string
	if err := pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient := pachClient.WithCtx(ctx)
		var iterators []Iterator
		for _, input := range inputs {
			fileSetID, err := Create(pachClient, taskDoer, input)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, fileSetID); err != nil {
				return err
			}
			iterators = append(iterators, NewFileSetIterator(pachClient, fileSetID, nil))
		}
		filesets, err := computeDatumKeyFilesets(pachClient, renewer, iterators, false)
		if err != nil {
			return err
		}
		var dits []Iterator
		for _, fs := range filesets {
			dits = append(dits, newFileSetMultiIterator(pachClient, fs))
		}
		outputFileSetID, err = WithCreateFileSet(pachClient, "pachyderm-datums-cross", func(s *Set) error {
			return mergeByKey(dits, existingMetaHash, func(metas []*Meta) error {
				var allInputs []*common.Input
				for _, m := range metas {
					allInputs = append(allInputs, m.Inputs...)
				}
				return s.UploadMeta(&Meta{Inputs: allInputs})
			})
		})
		return err
	}); err != nil {
		return "", err
	}
	return outputFileSetID, nil
}

func createCron(pachClient *client.APIClient, taskDoer task.Doer, input *pps.CronInput) (string, error) {
	return createPFS(pachClient, taskDoer, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

func serializePFSTask(task *PFSTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializePFSTaskResult(taskAny *types.Any) (*PFSTaskResult, error) {
	task := &PFSTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCrossTask(task *CrossTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeCrossTaskResult(taskAny *types.Any) (*CrossTaskResult, error) {
	task := &CrossTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}
