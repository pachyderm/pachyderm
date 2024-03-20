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
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if err := test.want.ValidateAll(); err != nil {
				t.Fatalf("test case 'want' fails validation: %v", err)
			}
			got, err := parseEditMetadataCmdline(test.args)
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

func TestEditMetadata_Empty(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl edit metadata
`).Run())
}

func TestEditMetadata_Add(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl inspect project default
		pachctl edit metadata project default add key1=value1
		pachctl inspect project default --raw | match '"key1":[[:space:]]+"value1"'
`).Run())
}

func TestEditMetadata_All(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutil.PachctlBashCmdCtx(ctx, t, c, `
		pachctl inspect project default
		pachctl edit metadata \
			project default set '{"key":"value","key2":"value"}' \
			project default add key3=value3 \
			project default delete key \
			project default edit key2=value2
		pachctl inspect project default --raw | match '"key2":[[:space:]]+"value2"'
`).Run())
}
