package datum

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	glob "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

func IsTaskResult(output *types.Any) bool {
	return types.Is(output, &PFSTaskResult{}) || types.Is(output, &CrossTaskResult{}) || types.Is(output, &ComposeTaskResult{})
}

func TaskResultFileSets(output *types.Any) ([]string, error) {
	switch {
	case types.Is(output, &PFSTaskResult{}):
		taskResult, err := deserializePFSTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	case types.Is(output, &CrossTaskResult{}):
		taskResult, err := deserializeCrossTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	case types.Is(output, &ComposeTaskResult{}):
		taskResult, err := deserializeComposeTaskResult(output)
		if err != nil {
			return nil, err
		}
		return []string{taskResult.FileSetId}, nil
	default:
		return nil, errors.Errorf("unrecognized any type (%v) in datum worker", output.TypeUrl)
	}
}

func IsTask(input *types.Any) bool {
	return types.Is(input, &PFSTask{}) || types.Is(input, &CrossTask{}) || types.Is(input, &ComposeTask{})
}

func ProcessTask(pachClient *client.APIClient, input *types.Any) (*types.Any, error) {
	switch {
	case types.Is(input, &PFSTask{}):
		task, err := deserializePFSTask(input)
		if err != nil {
			return nil, err
		}
		return processPFSTask(pachClient, task)
	case types.Is(input, &CrossTask{}):
		task, err := deserializeCrossTask(input)
		if err != nil {
			return nil, err
		}
		return processCrossTask(pachClient, task)
	case types.Is(input, &ComposeTask{}):
		task, err := deserializeComposeTask(input)
		if err != nil {
			return nil, err
		}
		return processComposeTask(pachClient, task)
	default:
		return nil, errors.Errorf("unrecognized any type (%v) in datum worker", input.TypeUrl)
	}
}

func processPFSTask(pachClient *client.APIClient, task *PFSTask) (*types.Any, error) {
	fileSetID, err := WithCreateFileSet(pachClient, "pachyderm-datums-pfs", func(s *Set) error {
		commit := client.NewCommit(task.Input.Repo, task.Input.Branch, task.Input.Commit)
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
		return clientsdk.ForEachGlobFile(client, func(fi *pfs.FileInfo) error {
			g := glob.MustCompile(pfsfile.CleanPath(task.Input.Glob), '/')
			// Remove the trailing slash to support glob replace on directory paths.
			p := strings.TrimRight(fi.File.Path, "/")
			joinOn := g.Replace(p, task.Input.JoinOn)
			groupBy := g.Replace(p, task.Input.GroupBy)
			meta := &Meta{
				Inputs: []*common.Input{
					&common.Input{
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

func processCrossTask(pachClient *client.APIClient, task *CrossTask) (*types.Any, error) {
	index := task.BaseIndex
	fileSetID, err := WithCreateFileSet(pachClient, "pachyderm-datums-cross", func(s *Set) error {
		iterators := []Iterator{NewFileSetIterator(pachClient, task.BaseFileSetId, task.BaseFileSetPathRange)}
		for _, fileSetID := range task.FileSetIds {
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

func processComposeTask(pachClient *client.APIClient, task *ComposeTask) (*types.Any, error) {
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

func deserializePFSTask(taskAny *types.Any) (*PFSTask, error) {
	task := &PFSTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializePFSTaskResult(task *PFSTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeCrossTask(taskAny *types.Any) (*CrossTask, error) {
	task := &CrossTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeCrossTaskResult(task *CrossTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeComposeTask(taskAny *types.Any) (*ComposeTask, error) {
	task := &ComposeTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeComposeTaskResult(task *ComposeTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}
