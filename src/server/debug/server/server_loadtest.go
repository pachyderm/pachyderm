package server

import (
	"context"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"sigs.k8s.io/yaml"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	srvpfs "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// RunLoadTest implements the pfs.RunLoadTest RPC
func (a *debugServer) RunPFSLoadTest(ctx context.Context, req *debug.RunPFSLoadTestRequest) (_ *debug.RunPFSLoadTestResponse, retErr error) {
	pachClient := a.env.GetPachClient(ctx)
	taskService := a.env.TaskService
	var project string
	repo := "load_test"
	if req.Branch != nil {
		project = req.Branch.Repo.Project.GetName()
		repo = req.Branch.Repo.Name
	}
	if err := pachClient.CreateRepo(project, repo); err != nil && !srvpfs.IsRepoExistsErr(err) {
		return nil, err
	}
	branch := uuid.New()
	if req.Branch != nil {
		branch = req.Branch.Name
	}
	if err := pachClient.CreateBranch(project, repo, branch, "", "", nil); err != nil {
		return nil, err
	}
	seed := time.Now().UTC().UnixNano()
	if req.Seed > 0 {
		seed = req.Seed
	}
	resp := &debug.RunPFSLoadTestResponse{
		Spec:   req.Spec,
		Branch: client.NewBranch(req.Branch.GetRepo().GetProject().GetName(), repo, branch),
		Seed:   seed,
	}
	start := time.Now()
	var err error
	resp.StateId, err = a.runLoadTest(pachClient, taskService, resp.Branch, req.Spec, seed, req.StateId)
	if err != nil {
		resp.Error = err.Error()
	}
	resp.Duration = durationpb.New(time.Since(start))
	return resp, nil
}

func (a *debugServer) runLoadTest(pachClient *client.APIClient, taskService task.Service, branch *pfs.Branch, specStr string, seed int64, stateID string) (string, error) {
	jsonBytes, err := yaml.YAMLToJSON([]byte(specStr))
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	spec := &pfsload.CommitSpec{}
	if err := protojson.Unmarshal(jsonBytes, spec); err != nil {
		return "", errors.Wrap(err, "unmarshal CommitSpec")
	}
	return pfsload.Commit(pachClient, taskService, branch, spec, seed, stateID)
}

func (a *debugServer) RunPFSLoadTestDefault(ctx context.Context, _ *emptypb.Empty) (resp *debug.RunPFSLoadTestResponse, retErr error) {
	for _, spec := range defaultLoadSpecs {
		var err error
		resp, err = a.RunPFSLoadTest(ctx, &debug.RunPFSLoadTestRequest{
			Spec: spec,
		})
		if err != nil {
			return nil, err
		}
		if resp.Error != "" {
			return resp, nil
		}
	}
	return resp, nil
}

var defaultLoadSpecs = []string{`
count: 3
modifications:
  - count: 5
    putFile:
      count: 5
      source: "random"
validator: {}
fileSources:
  - name: "random"
    random:
      directory:
        depth:
          min: 0
          max: 3
        run: 3
      sizes:
        - min: 1000
          max: 10000
          prob: 30
        - min: 10000
          max: 100000
          prob: 30
        - min: 1000000
          max: 10000000
          prob: 30
        - min: 10000000
          max: 100000000
          prob: 10
`, `
count: 3
modifications:
  - count: 5
    putFile:
      count: 10000
      source: "random"
validator: {}
fileSources:
  - name: "random"
    random:
      sizes:
        - min: 100
          max: 1000
          prob: 100
`, `
count: 3
modifications:
  - count: 5
    putFile:
      count: 1
      source: "random"
validator: {}
fileSources:
  - name: "random"
    random:
      sizes:
        - min: 10000000
          max: 100000000
          prob: 100
`}
