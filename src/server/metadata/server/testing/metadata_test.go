package testing

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestRealEnv(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	if _, err := env.PachClient.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestEditMetadata(t *testing.T) {
	c := pachd.NewTestPachd(t)
	ctx := c.Ctx()

	// Setup.
	project := &pfs.Project{
		Name: "foo",
	}
	if _, err := c.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{
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
	if _, err := c.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{
		Repo: repo,
	}); err != nil {
		t.Fatalf("create repo foo/test: %v", err)
	}

	branch := &pfs.Branch{
		Repo: repo,
		Name: "master",
	}
	if _, err := c.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{
		Branch: branch,
	}); err != nil {
		t.Fatalf("create branch foo/test@master: %v", err)
	}
	branchPicker := &pfs.BranchPicker{
		Picker: &pfs.BranchPicker_Name{
			Name: &pfs.BranchPicker_BranchName{
				Name: "master",
				Repo: &pfs.RepoPicker{
					Picker: &pfs.RepoPicker_Name{
						Name: &pfs.RepoPicker_RepoName{
							Name:    "test",
							Type:    "user",
							Project: projectPicker,
						},
					},
				},
			},
		},
	}

	commit := &pfs.Commit{
		Repo: repo,
		Branch: &pfs.Branch{
			Repo: repo,
			Name: "master",
		},
	}
	if err := c.PutFile(commit, "text.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("put file: %v", err)
	}
	commitPicker := &pfs.CommitPicker{
		Picker: &pfs.CommitPicker_BranchHead{
			BranchHead: branchPicker,
		},
	}

	t.Run("project", func(t *testing.T) {
		if _, err := c.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Project{
						Project: projectPicker,
					},
					Op: &metadata.Edit_AddKey_{
						AddKey: &metadata.Edit_AddKey{
							Key:   "key",
							Value: "value",
						},
					},
				},
				{
					Target: &metadata.Edit_Project{
						Project: &pfs.ProjectPicker{
							Picker: &pfs.ProjectPicker_Name{
								Name: "foo",
							},
						},
					},
					Op: &metadata.Edit_AddKey_{
						AddKey: &metadata.Edit_AddKey{
							Key:   "key2",
							Value: "value2",
						},
					},
				},
			},
		}); err != nil {
			t.Errorf("edit metadata: %v", err)
		}
		want := &pfs.ProjectInfo{
			Project: &pfs.Project{
				Name: "foo",
			},
			Description: "foo project",
			Metadata:    map[string]string{"key": "value", "key2": "value2"},
		}
		got, err := c.PfsAPIClient.InspectProject(ctx, &pfs.InspectProjectRequest{
			Project: want.Project,
		})
		if err != nil {
			t.Errorf("inspect project: %v", err)
		}
		got.CreatedAt = nil
		if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
			t.Errorf("project foo (-want +got):\n%s", diff)
		}
	})
	t.Run("commit", func(t *testing.T) {
		if _, err := c.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Commit{
						Commit: commitPicker,
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
		gotCommit, err := c.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: commit,
		})
		if err != nil {
			t.Errorf("inspect commit: %v", err)
		}
		if diff := cmp.Diff(want, gotCommit.GetMetadata()); diff != "" {
			t.Errorf("commit default/test@master (-want +got):\n%s", diff)
		}
	})

	t.Run("branch", func(t *testing.T) {
		if _, err := c.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Branch{
						Branch: branchPicker,
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
		gotBranch, err := c.PfsAPIClient.InspectBranch(ctx, &pfs.InspectBranchRequest{
			Branch: branch,
		})
		if err != nil {
			t.Errorf("inspect branch: %v", err)
		}
		if diff := cmp.Diff(want, gotBranch.GetMetadata()); diff != "" {
			t.Errorf("branch default/test@master (-want +got):\n%s", diff)
		}
	})

	t.Run("multi_failed", func(t *testing.T) {
		_, err := c.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Project{
						Project: projectPicker,
					},
					Op: &metadata.Edit_AddKey_{
						AddKey: &metadata.Edit_AddKey{
							Key:   "key",
							Value: "value",
						},
					},
				},
				{
					Target: &metadata.Edit_Branch{
						Branch: branchPicker,
					},
					Op: &metadata.Edit_AddKey_{
						AddKey: &metadata.Edit_AddKey{
							Key:   "key",
							Value: "value",
						},
					},
				},
				{
					Target: &metadata.Edit_Commit{
						Commit: commitPicker,
					},
					Op: &metadata.Edit_AddKey_{
						AddKey: &metadata.Edit_AddKey{
							Key:   "key",
							Value: "value",
						},
					},
				},
			},
		})
		if err == nil {
			t.Error("unexpected success")
		}
		want := "/(?ms:" +
			"FailedPrecondition.*" +
			"edit #0:.*already exists; use edit_key instead$" +
			".*^edit #1:.*already exists; use edit_key instead$" +
			".*^edit #2:.*already exists; use edit_key instead$" +
			")/"
		require.NoDiff(t, want, err.Error(), []cmp.Option{cmputil.RegexpStrings()})
	})
}
