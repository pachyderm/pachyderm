package cmds

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseEditMetadataCmdline(t *testing.T) {
	testData := []struct {
		name    string
		args    []string
		want    *metadata.EditMetadataRequest
		wantErr bool
	}{
		{
			name: "no args",
			want: &metadata.EditMetadataRequest{},
		},
		{
			name:    "too few args",
			args:    []string{"", "", ""},
			wantErr: true,
		},
		{
			name: "set project",
			args: []string{"project", "default", "set", `{"key":"value", "key2":"value2"}`},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_Replace_{
							Replace: &metadata.Edit_Replace{
								Replacement: map[string]string{
									"key":  "value",
									"key2": "value2",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "edit project",
			args: []string{"project", "default", "edit", "key=value"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_EditKey_{
							EditKey: &metadata.Edit_EditKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
				},
			},
		},
		{
			name: "add to project",
			args: []string{"project", "default", "add", "key=value"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_AddKey_{
							AddKey: &metadata.Edit_AddKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
				},
			},
		},
		{
			name: "delete from project",
			args: []string{"project", "default", "delete", "key"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_DeleteKey_{
							DeleteKey: &metadata.Edit_DeleteKey{
								Key: "key",
							},
						},
					},
				},
			},
		},
		{
			name: "do everything to a project",
			args: []string{
				"project", "default", "set", `{"key":"value", "key2":"value2"}`,
				"project", "default", "add", "key3=value",
				"project", "default", "edit", "key=new",
				"project", "default", "delete", "key",
			},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_Replace_{
							Replace: &metadata.Edit_Replace{
								Replacement: map[string]string{
									"key":  "value",
									"key2": "value2",
								},
							},
						},
					},
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_AddKey_{
							AddKey: &metadata.Edit_AddKey{
								Key:   "key3",
								Value: "value",
							},
						},
					},
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_EditKey_{
							EditKey: &metadata.Edit_EditKey{
								Key:   "key",
								Value: "new",
							},
						},
					},
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_DeleteKey_{
							DeleteKey: &metadata.Edit_DeleteKey{
								Key: "key",
							},
						},
					},
				},
			},
		},
		{
			name: "set metadata for a project",
			args: []string{"project", "default", "set", "key=new"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Project{
							Project: &pfs.ProjectPicker{
								Picker: &pfs.ProjectPicker_Name{
									Name: "default",
								},
							},
						},
						Op: &metadata.Edit_Replace_{
							Replace: &metadata.Edit_Replace{
								Replacement: map[string]string{"key": "new"},
							},
						},
					},
				},
			},
		},
		{
			name:    "wrong syntax for set",
			args:    []string{"project", "default", "set", ";key=value"},
			wantErr: true,
		},
		{
			name: "edit a commit",
			args: []string{"commit", "images@master", "edit", "key=value",
				"commit", "images@44444444444444444444444444444444", "edit", "key=value"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Commit{
							Commit: &pfs.CommitPicker{
								Picker: &pfs.CommitPicker_BranchHead{
									BranchHead: &pfs.BranchPicker{
										Picker: &pfs.BranchPicker_Name{
											Name: &pfs.BranchPicker_BranchName{
												Repo: &pfs.RepoPicker{
													Picker: &pfs.RepoPicker_Name{
														Name: &pfs.RepoPicker_RepoName{
															Project: &pfs.ProjectPicker{
																Picker: &pfs.ProjectPicker_Name{
																	Name: "the_default_project",
																},
															},
															Name: "images",
															Type: "user",
														},
													},
												},
												Name: "master",
											},
										},
									},
								},
							},
						},
						Op: &metadata.Edit_EditKey_{
							EditKey: &metadata.Edit_EditKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
					{
						Target: &metadata.Edit_Commit{
							Commit: &pfs.CommitPicker{
								Picker: &pfs.CommitPicker_Id{
									Id: &pfs.CommitPicker_CommitByGlobalId{
										Repo: &pfs.RepoPicker{
											Picker: &pfs.RepoPicker_Name{
												Name: &pfs.RepoPicker_RepoName{
													Project: &pfs.ProjectPicker{
														Picker: &pfs.ProjectPicker_Name{
															Name: "the_default_project",
														},
													},
													Name: "images",
													Type: "user",
												},
											},
										},
										Id: "44444444444444444444444444444444",
									},
								},
							},
						},
						Op: &metadata.Edit_EditKey_{
							EditKey: &metadata.Edit_EditKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
				},
			},
		},
		{
			name: "add to a branch",
			args: []string{"branch", "test@master", "add", "key=value"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Branch{
							Branch: &pfs.BranchPicker{
								Picker: &pfs.BranchPicker_Name{
									Name: &pfs.BranchPicker_BranchName{
										Repo: &pfs.RepoPicker{
											Picker: &pfs.RepoPicker_Name{
												Name: &pfs.RepoPicker_RepoName{
													Project: &pfs.ProjectPicker{
														Picker: &pfs.ProjectPicker_Name{
															Name: "the_default_project",
														},
													},
													Name: "test",
													Type: "user",
												},
											},
										},
										Name: "master",
									},
								},
							},
						},
						Op: &metadata.Edit_AddKey_{
							AddKey: &metadata.Edit_AddKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
				},
			},
		},
		{
			name: "add to a user repo",
			args: []string{"repo", "test", "add", "key=value"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Repo{
							Repo: &pfs.RepoPicker{
								Picker: &pfs.RepoPicker_Name{
									Name: &pfs.RepoPicker_RepoName{
										Project: &pfs.ProjectPicker{
											Picker: &pfs.ProjectPicker_Name{
												Name: "the_default_project",
											},
										},
										Name: "test",
										Type: "user",
									},
								},
							},
						},
						Op: &metadata.Edit_AddKey_{
							AddKey: &metadata.Edit_AddKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
				},
			},
		},
		{
			name: "add to a spec repo",
			args: []string{"repo", "default/test.spec", "add", "key=value"},
			want: &metadata.EditMetadataRequest{
				Edits: []*metadata.Edit{
					{
						Target: &metadata.Edit_Repo{
							Repo: &pfs.RepoPicker{
								Picker: &pfs.RepoPicker_Name{
									Name: &pfs.RepoPicker_RepoName{
										Project: &pfs.ProjectPicker{
											Picker: &pfs.ProjectPicker_Name{
												Name: "default",
											},
										},
										Name: "test",
										Type: "spec",
									},
								},
							},
						},
						Op: &metadata.Edit_AddKey_{
							AddKey: &metadata.Edit_AddKey{
								Key:   "key",
								Value: "value",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if err := test.want.ValidateAll(); err != nil {
				t.Fatalf("test case 'want' fails validation: %v", err)
			}
			got, err := parseEditMetadataCmdline(test.args, "the_default_project")
			if test.want == nil {
				test.want = &metadata.EditMetadataRequest{}
			}
			if got == nil {
				got = &metadata.EditMetadataRequest{}
			}
			if diff := cmp.Diff(test.want, got, protocmp.Transform(), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("request (-want +got):\n%s", diff)
			}
			if err == nil && test.wantErr {
				t.Error("expected error, but got success")
				return
			}
			if err != nil && !test.wantErr {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if err := got.ValidateAll(); err != nil {
				t.Fatalf("request fails validation: %v", err)
			}
		})
	}
}

func TestEditMetadata(t *testing.T) {
	c := pachd.NewTestPachd(t)
	testData := []struct {
		name string
		code string
	}{
		{
			name: "empty",
			code: "pachctl edit metadata",
		},
		{
			name: "project",
			code: `
				pachctl edit metadata project default add key1=value1
				pachctl inspect project default --raw | match '"key1":[[:space:]]+"value1"'
			`,
		},
		{
			name: "commit",
			code: `
				pachctl edit metadata commit test@master set '{"key":"value"}'
				pachctl inspect commit test@master --raw | match '"key":[[:space:]]+"value"'
			`,
		},
		{
			name: "branch",
			code: `
				pachctl edit metadata branch test@master set '{"key":"value"}'
				pachctl inspect branch test@master --raw | match '"key":[[:space:]]+"value"'
			`,
		},
		{
			name: "repo",
			code: `
				pachctl edit metadata repo test set '{"key":"value"}'
				pachctl inspect repo test --raw | match '"key":[[:space:]]+"value"'
			`,
		},
		{
			name: "cluster",
			code: `
				pachctl edit metadata cluster . set '{"key":"value"}'
				pachctl inspect cluster --raw | match '"key":[[:space:]]+"value"'
			`,
		},
		{
			name: "all",
			code: `
				pachctl inspect project default
				pachctl edit metadata \
					project default set '{"key":"value","key2":"value"}' \
					project default add key3=value3 \
					project default delete key \
					project default edit key2=value2
				pachctl inspect project default --raw | match '"key2":[[:space:]]+"value2"'
			`,
		},
		{
			name: "non-default project",
			code: `
				pachctl create project project
				pachctl config update context --project project
				pachctl create repo test
				echo 'hi' | pachctl put file test@master:/hi.txt

				pachctl edit metadata commit project/test@master add myproject=project \
						      commit default/test@master add myproject=default \
						      commit test@master add implied=true \
						      branch default/test@master add mybranchproject=default \
						      branch project/test@master add mybranchproject=project \
						      branch test@master add branchimplied=true \
						      repo default/test add myrepoproject=default \
						      repo project/test add myrepoproject=project \
						      repo test add repoimplied=true
				pachctl inspect commit test@master --raw --project=default |
					match '"myproject":[[:space:]]+"default"'
				pachctl inspect commit test@master --raw --project=project |
					match '"implied":[[:space:]]+"true"' |
					match '"myproject":[[:space:]]+"project"'
				pachctl inspect branch test@master --raw --project=default |
					match '"mybranchproject":[[:space:]]+"default"'
				pachctl inspect branch test@master --raw --project=project |
					match '"branchimplied":[[:space:]]+"true"' |
					match '"mybranchproject":[[:space:]]+"project"'
				pachctl inspect repo test --raw --project=default |
					match '"myrepoproject":[[:space:]]+"default"'
				pachctl inspect repo test --raw --project=project |
					match '"repoimplied":[[:space:]]+"true"' |
					match '"myrepoproject":[[:space:]]+"project"'
			`,
		},
	}
	ctx := pctx.TestContext(t)
	setup := `
		pachctl create repo test
		echo 'hi' | pachctl put file test@master:/hi.txt
	`
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, setup).Run())
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, test.code).Run())
		})
	}
}
