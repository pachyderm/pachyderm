package testing

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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
	testutil.ActivateAuthClient(t, c, strconv.Itoa(int(c.GetAddress().Port)))
	root := testutil.AuthenticateClient(t, c, auth.RootUser)
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
	repoPicker := &pfs.RepoPicker{
		Picker: &pfs.RepoPicker_Name{
			Name: &pfs.RepoPicker_RepoName{
				Project: projectPicker,
				Name:    repo.GetName(),
				Type:    repo.GetType(),
			},
		},
	}

	branch := &pfs.Branch{
		Repo: repo,
		Name: "master",
	}
	if _, err := root.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{
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
	if err := root.PutFile(commit, "text.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("put file: %v", err)
	}
	commitPicker := &pfs.CommitPicker{
		Picker: &pfs.CommitPicker_BranchHead{
			BranchHead: branchPicker,
		},
	}

	// Grant auth permissions.
	require.NoError(t, root.ModifyRepoRoleBinding(root.Ctx(), "foo", "test", "robot:alice", []string{auth.RepoWriterRole}))

	// Tests follow.
	t.Run("project", func(t *testing.T) {
		if _, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
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
		want := map[string]string{"key": "value", "key2": "value2"}
		gotProject, err := root.PfsAPIClient.InspectProject(ctx, &pfs.InspectProjectRequest{
			Project: project,
		})
		if err != nil {
			t.Errorf("inspect project: %v", err)
		}
		if diff := cmp.Diff(want, gotProject.Metadata, protocmp.Transform()); diff != "" {
			t.Errorf("project foo (-want +got):\n%s", diff)
		}
	})

	t.Run("commit", func(t *testing.T) {
		if _, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
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
		gotCommit, err := root.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
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
		if _, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
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
		gotBranch, err := root.PfsAPIClient.InspectBranch(ctx, &pfs.InspectBranchRequest{
			Branch: branch,
		})
		if err != nil {
			t.Errorf("inspect branch: %v", err)
		}
		if diff := cmp.Diff(want, gotBranch.GetMetadata()); diff != "" {
			t.Errorf("branch default/test@master (-want +got):\n%s", diff)
		}
	})

	t.Run("repo", func(t *testing.T) {
		if _, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Repo{
						Repo: repoPicker,
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
		got, err := root.PfsAPIClient.InspectRepo(ctx, &pfs.InspectRepoRequest{
			Repo: repo,
		})
		if err != nil {
			t.Errorf("inspect repo: %v", err)
		}
		if diff := cmp.Diff(want, got.GetMetadata()); diff != "" {
			t.Errorf("repo default/test (-want +got):\n%s", diff)
		}
	})

	t.Run("multi_success", func(t *testing.T) {
		op := &metadata.Edit_EditKey_{
			EditKey: &metadata.Edit_EditKey{
				Key:   "key",
				Value: "value2",
			},
		}
		_, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Project{
						Project: projectPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Branch{
						Branch: branchPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Commit{
						Commit: commitPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Repo{
						Repo: repoPicker,
					},
					Op: op,
				},
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("multi_failed", func(t *testing.T) {
		op := &metadata.Edit_AddKey_{
			AddKey: &metadata.Edit_AddKey{
				Key:   "key",
				Value: "value",
			},
		}
		_, err := root.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
			Edits: []*metadata.Edit{
				{
					Target: &metadata.Edit_Project{
						Project: projectPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Branch{
						Branch: branchPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Commit{
						Commit: commitPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Repo{
						Repo: repoPicker,
					},
					Op: op,
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
			".*^edit #3:.*already exists; use edit_key instead$" +
			")/"
		require.NoDiff(t, want, err.Error(), []cmp.Option{cmputil.RegexpStrings()})
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
					Target: &metadata.Edit_Project{
						Project: projectPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Branch{
						Branch: branchPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Commit{
						Commit: commitPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Repo{
						Repo: repoPicker,
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
					Target: &metadata.Edit_Project{
						Project: projectPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Branch{
						Branch: branchPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Commit{
						Commit: commitPicker,
					},
					Op: op,
				},
				{
					Target: &metadata.Edit_Repo{
						Repo: repoPicker,
					},
					Op: op,
				},
			},
		})
		if err == nil {
			t.Error("unexpected success")
		}
		want := []string{
			`/PermissionDenied.*edit #1: check permissions on branch.*robot:mallory is not authorized/`,
			`/edit #3: check permissions on repo.*robot:mallory is not authorized/`,
		}
		require.NoDiff(t, want, strings.Split(err.Error(), "\n"), []cmp.Option{cmputil.RegexpStrings()})
	})
}
