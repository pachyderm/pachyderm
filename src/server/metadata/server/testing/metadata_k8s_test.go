//go:build k8s

package testing

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestPipeline(t *testing.T) {
	root := pachd.NewTestPachd(t, pachd.ActivateAuthOption(""))
	ctx := root.Ctx()

	alice := testutil.AuthenticateClient(t, root, "alice")     // alice has the necessary permissions
	mallory := testutil.AuthenticateClient(t, root, "mallory") // mallory does not have any permissions

	// Setup.
	project := &pfs.Project{
		Name: "foo",
	}
	if _, err := root.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{
		Project:     project,
		Description: "foo project",
	}); err != nil {
		t.Fatalf("create project foo: %v", err)
	}
	projectPicker := &pfs.ProjectPicker{
		Picker: &pfs.ProjectPicker_Name{
			Name: "foo",
		},
	}

	repo := &pfs.Repo{
		Name:    "test",
		Type:    "user",
		Project: project,
	}
	if _, err := root.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{
		Repo: repo,
	}); err != nil {
		t.Fatalf("create repo foo/test: %v", err)
	}
	// repoPicker := &pfs.RepoPicker{
	// 	Picker: &pfs.RepoPicker_Name{
	// 		Name: &pfs.RepoPicker_RepoName{
	// 			Project: projectPicker,
	// 			Name:    repo.GetName(),
	// 			Type:    repo.GetType(),
	// 		},
	// 	},
	// }

	branch := &pfs.Branch{
		Repo: repo,
		Name: "master",
	}
	if _, err := root.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{
		Branch: branch,
	}); err != nil {
		t.Fatalf("create branch foo/test@master: %v", err)
	}
	// branchPicker := &pfs.BranchPicker{
	// 	Picker: &pfs.BranchPicker_Name{
	// 		Name: &pfs.BranchPicker_BranchName{
	// 			Name: "master",
	// 			Repo: &pfs.RepoPicker{
	// 				Picker: &pfs.RepoPicker_Name{
	// 					Name: &pfs.RepoPicker_RepoName{
	// 						Name:    "test",
	// 						Type:    "user",
	// 						Project: projectPicker,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	commit := &pfs.Commit{
		Repo: repo,
		Branch: &pfs.Branch{
			Repo: repo,
			Name: "master",
		},
	}
	if err := root.PutFile(commit, "text.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("put file: %v", err)
	}
	// commitPicker := &pfs.CommitPicker{
	// 	Picker: &pfs.CommitPicker_BranchHead{
	// 		BranchHead: branchPicker,
	// 	},
	// }

	var (
		pipeline = &pps.Pipeline{
			Project: project,
			Name:    "test-pipeline",
		}
		pipelinePicker = &pps.PipelinePicker{
			Picker: &pps.PipelinePicker_Name{
				Name: &pps.PipelinePicker_RepoName{
					Project: projectPicker,
					Name:    "test-pipeline",
				},
			},
		}
	)
	if _, err := root.CreatePipelineV2(ctx, &pps.CreatePipelineV2Request{
		CreatePipelineRequestJson: `{
  "pipeline": {
    "project": {"name": "foo"},
    "name": "test-pipeline"
  },
  "transform": {
    "cmd": [
	"sh", "-c", "ls -R /pfs > /pfs/out/output.txt"
    ]
  },
  "input": {
    "cross": [
      {
        "pfs": {
          "repo": "test",
          "glob": "/",
          "name": "a"
        }
      },
      {
        "pfs": {
          "repo": "test",
          "glob": "/",
          "name": "b"
        }
      }
    ]
  },
  "resource_requests": {
    "cpu": 0.5
  },
  "autoscaling": false
}`,
	}); err != nil {
		t.Fatalf("create pipeline foo/test-pipeline: %v", err)
	}

	// Grant auth permissions.
	require.NoError(t, root.ModifyRepoRoleBinding(ctx, "foo", "test", "robot:alice", []string{auth.RepoWriterRole}))
	require.NoError(t, root.ModifyRepoRoleBinding(ctx, "foo", "test-pipeline", "robot:alice", []string{auth.RepoWriterRole}))

	t.Run("pipeline", func(t *testing.T) {
		if _, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Pipeline{
						Pipeline: pipelinePicker,
					},
					Op: &metadata.Edit_AddKey_{
						AddKey: &metadata.Edit_AddKey{
							Key:   "key",
							Value: "value",
						},
					},
				},
			},
		}); err != nil {
			t.Errorf("edit metadata: %v", err)
		}
		want := map[string]string{"key": "value"}
		got, err := root.PpsAPIClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{
			Pipeline: pipeline,
		})
		if err != nil {
			t.Errorf("inspect repo: %v", err)
		}
		if diff := cmp.Diff(want, got.GetMetadata()); diff != "" {
			t.Errorf("repo default/test (-want +got):\n%s", diff)
		}
	})

	t.Run("auth_allowed", func(t *testing.T) {
		op := &metadata.Edit_EditKey_{
			EditKey: &metadata.Edit_EditKey{
				Key:   "key",
				Value: "value2",
			},
		}
		_, err := alice.MetadataClient.EditMetadata(alice.Ctx(), &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Pipeline{
						Pipeline: pipelinePicker,
					},
					Op: op,
				},
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("auth_not_allowed", func(t *testing.T) {
		op := &metadata.Edit_EditKey_{
			EditKey: &metadata.Edit_EditKey{
				Key:   "key",
				Value: "value2",
			},
		}
		_, err := mallory.MetadataClient.EditMetadata(mallory.Ctx(), &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Pipeline{
						Pipeline: pipelinePicker,
					},
					Op: op,
				},
			},
		})
		if err == nil {
			t.Error("unexpected success")
		}
		require.ErrorContains(t, err, `rpc error: code = PermissionDenied`)
	})
}
