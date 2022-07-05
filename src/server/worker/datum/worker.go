package datum

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	glob "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

func IsTask(input *types.Any) bool {
	switch {
	case types.Is(input, &PFSTask{}):
		return true
	default:
		return false
	}
}

func ProcessTask(pachClient *client.APIClient, input *types.Any) (*types.Any, error) {
	switch {
	case types.Is(input, &PFSTask{}):
		task, err := deserializePFSTask(input)
		if err != nil {
			return nil, err
		}
		return processPFSTask(pachClient, task)
	default:
		return nil, errors.Errorf("unrecognized any type (%v) in datum worker", input.TypeUrl)
	}
}

func processPFSTask(pachClient *client.APIClient, task *PFSTask) (*types.Any, error) {
	fileSetID, err := withDatumFileSet(pachClient, func(s *Set) error {
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
			return err
		}
		return clientsdk.ForEachGlobFile(client, func(fi *pfs.FileInfo) error {
			g := glob.MustCompile(pfsfile.CleanPath(task.Input.Glob), '/')
			// Remove the trailing slash to support glob replace on directory paths.
			p := strings.TrimRight(fi.File.Path, "/")
			joinOn := g.Replace(p, task.Input.JoinOn)
			groupBy := g.Replace(p, task.Input.GroupBy)
			return s.UploadMeta(&Meta{
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
			})
		})
	})
	if err != nil {
		return nil, err
	}
	return serializePFSTaskResult(&PFSTaskResult{FileSetId: fileSetID})
}

func withDatumFileSet(pachClient *client.APIClient, cb func(*Set) error) (string, error) {
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		storageRoot := filepath.Join(os.TempDir(), "pachyderm-datums", uuid.NewWithoutDashes())
		return WithSet(nil, storageRoot, cb, WithMetaOutput(mf))
	})
	if err != nil {
		return "", err
	}
	return resp.FileSetId, nil
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
