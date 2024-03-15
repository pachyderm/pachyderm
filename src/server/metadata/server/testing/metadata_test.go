package testing

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
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

func TestEditProjectMetadata(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	ctx = c.Ctx()
	if _, err := c.PfsAPIClient.CreateProject(ctx, &pfs.CreateProjectRequest{
		Project: &pfs.Project{
			Name: "foo",
		},
		Description: "foo project",
	}); err != nil {
		t.Fatalf("create project foo: %v", err)
	}

	if _, err := c.MetadataClient.EditMetadata(ctx, &metadata.EditMetadataRequest{
		Edits: []*metadata.Edit{
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
}
