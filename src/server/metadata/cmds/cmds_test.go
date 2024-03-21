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
			name:    "wrong syntax for set",
			args:    []string{"project", "default", "set", "key=value"},
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
			name: "add",
			code: `
				pachctl inspect project default
				pachctl edit metadata project default add key1=value1
				pachctl inspect project default --raw | match '"key1":[[:space:]]+"value1"'
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
			name: "commit",
			code: `
				pachctl create repo test || true
				echo 'hi' | pachctl put file test@master:/hi.txt
				pachctl edit metadata commit test@master set '{"key":"value"}'
				pachctl inspect commit test@master --raw | match '"key":[[:space:]]+"value"'
			`,
		},
		{
			name: "non-default project",
			code: `
				pachctl create repo test || true
				echo 'hi' | pachctl put file test@master:/hi.txt
				
				pachctl create project project
				pachctl config update context --project project
				pachctl create repo test
				echo 'hi' | pachctl put file test@master:/hi.txt
				
				pachctl edit metadata commit project/test@master edit myproject=project \
						      commit default/test@master edit myproject=default \
						      commit test@master add implied=true
				pachctl inspect commit test@master --raw --project=default |
					match '"myproject":[[:space:]]+"default"'
				pachctl inspect commit test@master --raw --project=project |
					match '"implied":[[:space:]]+"true"' |
					match '"myproject":[[:space:]]+"project"'
			`,
		},
		{
			name: "branch",
			code: `
				pachctl create project branchtest
				pachctl config update context --project branchtest
				pachctl create repo testbranch
				pachctl create branch testbranch@foobar
				pachctl edit metadata branch testbranch@foobar add key=value \
					branch branchtest/testbranch@foobar add key2=value2
				pachctl inspect branch testbranch@foobar --raw |
					match '"key":[[:space:]]"value"' |
					match '"key2":[[:space:]]"value2"'
			`,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, test.code).Run())
		})
	}
}
