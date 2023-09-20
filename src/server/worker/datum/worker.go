package datum

import (
	"bytes"
	"encoding/hex"
	"strings"

	glob "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func IsTaskResult(output *anypb.Any) bool {
	return output.MessageIs(&PFSTaskResult{}) ||
		output.MessageIs(&CrossTaskResult{}) ||
		output.MessageIs(&KeyTaskResult{}) ||
		output.MessageIs(&MergeTaskResult{}) ||
		output.MessageIs(&ComposeTaskResult{})
}

func TaskResultFileSets(output *anypb.Any) ([]string, error) {
	switch {
	case output.MessageIs(&PFSTaskResult{}):
		taskResult, err := deserializePFSTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	case output.MessageIs(&CrossTaskResult{}):
		taskResult, err := deserializeCrossTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	case output.MessageIs(&KeyTaskResult{}):
		taskResult, err := deserializeKeyTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	case output.MessageIs(&MergeTaskResult{}):
		taskResult, err := deserializeMergeTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	case output.MessageIs(&ComposeTaskResult{}):
		taskResult, err := deserializeComposeTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	default:
		return nil, errors.Errorf("unrecognized any type (%v) in datum worker", output.TypeUrl)
	}
}

func IsTask(input *anypb.Any) bool {
	return input.MessageIs(&PFSTask{}) ||
		input.MessageIs(&CrossTask{}) ||
		input.MessageIs(&KeyTask{}) ||
		input.MessageIs(&MergeTask{}) ||
		input.MessageIs(&ComposeTask{})
}

func ProcessTask(pachClient *client.APIClient, input *anypb.Any) (*anypb.Any, error) {
	switch {
	case input.MessageIs(&PFSTask{}):
		task, err := deserializePFSTask(input)
		if err != nil {
			return nil, err
		}
		pachClient.SetAuthToken(task.AuthToken)
		return processPFSTask(pachClient, task)
	case input.MessageIs(&CrossTask{}):
		task, err := deserializeCrossTask(input)
		if err != nil {
			return nil, err
		}
		pachClient.SetAuthToken(task.AuthToken)
		return processCrossTask(pachClient, task)
	case input.MessageIs(&KeyTask{}):
		task, err := deserializeKeyTask(input)
		if err != nil {
			return nil, err
		}
		pachClient.SetAuthToken(task.AuthToken)
		return processKeyTask(pachClient, task)
	case input.MessageIs(&MergeTask{}):
		task, err := deserializeMergeTask(input)
		if err != nil {
			return nil, err
		}
		pachClient.SetAuthToken(task.AuthToken)
		return processMergeTask(pachClient, task)
	case input.MessageIs(&ComposeTask{}):
		task, err := deserializeComposeTask(input)
		if err != nil {
			return nil, err
		}
		pachClient.SetAuthToken(task.AuthToken)
		return processComposeTask(pachClient, task)
	default:
		return nil, errors.Errorf("unrecognized any type (%v) in datum worker", input.TypeUrl)
	}
}

func processPFSTask(pachClient *client.APIClient, task *PFSTask) (*anypb.Any, error) {
	fileSetID, err := WithCreateFileSet(pachClient, "pachyderm-datums-pfs", func(s *Set) error {
		commit := client.NewCommit(task.Input.Project, task.Input.Repo, task.Input.Branch, task.Input.Commit)
		client, err := pachClient.PfsAPIClient.GlobFile(
			pachClient.Ctx(),
			&pfs.GlobFileRequest{
				Commit:    commit,
				Pattern:   task.Input.Glob,
				PathRange: task.PathRange,
			},
		)
		if err != nil {
			return errors.EnsureStack(err)
		}
		index := task.BaseIndex
		return grpcutil.ForEach[*pfs.FileInfo](client, func(fi *pfs.FileInfo) error {
			g := glob.MustCompile(pfsfile.CleanPath(task.Input.Glob), '/')
			// Remove the trailing slash to support glob replace on directory paths.
			p := strings.TrimRight(fi.File.Path, "/")
			joinOn := g.Replace(p, task.Input.JoinOn)
			groupBy := g.Replace(p, task.Input.GroupBy)
			meta := &Meta{
				Inputs: []*common.Input{
					{
						FileInfo:   fi,
						JoinOn:     joinOn,
						OuterJoin:  task.Input.OuterJoin,
						GroupBy:    groupBy,
						Name:       task.Input.Name,
						Lazy:       task.Input.Lazy,
						Branch:     task.Input.Branch,
						EmptyFiles: task.Input.EmptyFiles,
						S3:         task.Input.S3,
					},
				},
				Index: index,
			}
			index++
			return s.UploadMeta(meta)
		})
	})
	if err != nil {
		return nil, err
	}
	return serializePFSTaskResult(&PFSTaskResult{FileSetId: fileSetID})
}

func processCrossTask(pachClient *client.APIClient, task *CrossTask) (*anypb.Any, error) {
	index := task.BaseIndex
	fileSetID, err := WithCreateFileSet(pachClient, "pachyderm-datums-cross", func(s *Set) error {
		var iterators []Iterator
		for i, fileSetID := range task.FileSetIds {
			if i == int(task.BaseFileSetIndex) {
				iterators = append(iterators, NewFileSetIterator(pachClient, fileSetID, task.BaseFileSetPathRange))
				continue
			}
			iterators = append(iterators, NewFileSetIterator(pachClient, fileSetID, nil))
		}
		return iterate(nil, iterators, func(meta *Meta) error {
			meta.Index = index
			index++
			return s.UploadMeta(meta)
		})
	})
	if err != nil {
		return nil, err
	}
	return serializeCrossTaskResult(&CrossTaskResult{FileSetId: fileSetID})
}

func iterate(crossInputs []*common.Input, iterators []Iterator, cb func(*Meta) error) error {
	if len(iterators) == 0 {
		return cb(&Meta{Inputs: crossInputs})
	}
	// TODO: Might want to exit fast for the zero datums case.
	err := iterators[0].Iterate(func(meta *Meta) error {
		return iterate(append(crossInputs, meta.Inputs...), iterators[1:], cb)
	})
	return errors.EnsureStack(err)
}

func processKeyTask(pachClient *client.APIClient, task *KeyTask) (*anypb.Any, error) {
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		fsi := NewFileSetIterator(pachClient, task.FileSetId, task.PathRange)
		keyHasher := pfs.NewHash()
		marshaller := protojson.MarshalOptions{}
		err := fsi.Iterate(func(meta *Meta) error {
			for _, input := range meta.Inputs {
				var rawKey string
				switch task.Type {
				case KeyTask_JOIN:
					rawKey = input.JoinOn
				case KeyTask_GROUP:
					rawKey = input.GroupBy
				default:
					return errors.Errorf("unrecognized key task type (%v)", task.Type)
				}
				// hash input keys to ensure consistently-shaped filepaths
				keyHasher.Reset()
				keyHasher.Write([]byte(rawKey))
				key := hex.EncodeToString(keyHasher.Sum(nil))
				out, err := marshaller.Marshal(input)
				if err != nil {
					return errors.Wrap(err, "marshalling input for key aggregation")
				}
				if err := mf.PutFile(key, bytes.NewReader(out), client.WithAppendPutFile()); err != nil {
					return errors.EnsureStack(err)
				}
			}
			return nil
		})
		return errors.EnsureStack(err)
	})
	if err != nil {
		return nil, err
	}
	return serializeKeyTaskResult(&KeyTaskResult{FileSetId: resp.FileSetId})
}

func processMergeTask(pachClient *client.APIClient, task *MergeTask) (*anypb.Any, error) {
	fileSetID, err := WithCreateFileSet(pachClient, "pachyderm-datums-merge", func(s *Set) error {
		var fsmis []Iterator
		for _, fileSetId := range task.FileSetIds {
			fsmis = append(fsmis, newFileSetMultiIterator(pachClient, fileSetId, task.PathRange))
		}
		switch task.Type {
		case MergeTask_JOIN:
			return mergeByKey(fsmis, existingMetaHash, func(metas []*Meta) error {
				var crossInputs [][]*common.Input
				for _, m := range metas {
					crossInputs = append(crossInputs, m.Inputs)
				}
				err := newCrossListIterator(crossInputs).Iterate(func(meta *Meta) error {
					if len(meta.Inputs) == len(task.FileSetIds) {
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
		case MergeTask_GROUP:
			return mergeByKey(fsmis, existingMetaHash, func(metas []*Meta) error {
				var allInputs []*common.Input
				for _, m := range metas {
					allInputs = append(allInputs, m.Inputs...)
				}
				return s.UploadMeta(&Meta{Inputs: allInputs})
			})
		default:
			return errors.Errorf("unrecognized merge task type (%v)", task.Type)
		}
	})
	if err != nil {
		return nil, err
	}
	return serializeMergeTaskResult(&MergeTaskResult{FileSetId: fileSetID})
}

func existingMetaHash(meta *Meta) string {
	return meta.Hash
}

func processComposeTask(pachClient *client.APIClient, task *ComposeTask) (*anypb.Any, error) {
	resp, err := pachClient.PfsAPIClient.ComposeFileSet(
		pachClient.Ctx(),
		&pfs.ComposeFileSetRequest{
			FileSetIds: task.FileSetIds,
			TtlSeconds: int64(common.TTL),
			Compact:    true,
		},
	)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return serializeComposeTaskResult(&ComposeTaskResult{FileSetId: resp.FileSetId})
}

func deserializePFSTask(taskAny *anypb.Any) (*PFSTask, error) {
	task := &PFSTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializePFSTaskResult(task *PFSTaskResult) (*anypb.Any, error) { return anypb.New(task) }

func deserializeCrossTask(taskAny *anypb.Any) (*CrossTask, error) {
	task := &CrossTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeCrossTaskResult(task *CrossTaskResult) (*anypb.Any, error) { return anypb.New(task) }

func deserializeKeyTask(taskAny *anypb.Any) (*KeyTask, error) {
	task := &KeyTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeKeyTaskResult(task *KeyTaskResult) (*anypb.Any, error) { return anypb.New(task) }

func deserializeMergeTask(taskAny *anypb.Any) (*MergeTask, error) {
	task := &MergeTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeMergeTaskResult(task *MergeTaskResult) (*anypb.Any, error) { return anypb.New(task) }

func deserializeComposeTask(taskAny *anypb.Any) (*ComposeTask, error) {
	task := &ComposeTask{}
	if err := taskAny.UnmarshalTo(task); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return task, nil
}

func serializeComposeTaskResult(task *ComposeTaskResult) (*anypb.Any, error) { return anypb.New(task) }
