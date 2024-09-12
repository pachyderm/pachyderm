package pfsdb_test

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	testProj     = "testproj"
	testProjDesc = "this is a test project"
)

var (
	testMetadata = map[string]string{"metadata": "is here"}
)

func TestCreateProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateProject(cbCtx, tx, createInfo), "should be able to create project")
		getInfo, err := pfsdb.GetProjectInfoByName(cbCtx, tx, testProj)
		require.NoError(t, err, "should be able to get a project")
		require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		return nil
	}))
	require.YesError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		err := pfsdb.CreateProject(cbCtx, tx, createInfo)
		require.YesError(t, err, "should not be able to create project again with same name")
		require.True(t, (&pfsdb.ProjectAlreadyExistsError{Name: testProj}).Is(err))
		return nil
	}), "double create should fail and result in rollback")
}

func TestDeleteProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, Metadata: testMetadata}
		require.NoError(t, pfsdb.CreateProject(cbCtx, tx, createInfo), "should be able to create project")
		require.NoError(t, pfsdb.DeleteProject(cbCtx, tx, createInfo.Project.Name), "should be able to delete project")
		_, err := pfsdb.GetProjectInfoByName(cbCtx, tx, testProj)
		require.YesError(t, err, "get project should not find row")
		targetErr := &pfsdb.ProjectNotFoundError{Name: testProj}
		require.True(t, targetErr.Is(err))
		err = pfsdb.DeleteProject(cbCtx, tx, createInfo.Project.Name)
		require.YesError(t, err, "double delete should be an error")
		require.True(t, targetErr.Is(err))
		return nil
	}))
}

func TestGetProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, Metadata: testMetadata, CreatedAt: timestamppb.Now()}
		require.NoError(t, pfsdb.CreateProject(cbCtx, tx, createInfo), "should be able to create project")
		// the 'default' project already exists with an ID 1.
		getInfo, err := pfsdb.GetProjectInfo(cbCtx, tx, 2)
		require.NoError(t, err, "should be able to get a project")
		require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		require.Equal(t, createInfo.Metadata, getInfo.Metadata)
		// validate GetProject.
		project, err := pfsdb.GetProject(cbCtx, tx, testProj)
		require.NoError(t, err, "should be able to get a project")
		require.Equal(t, createInfo.Project.Name, project.ProjectInfo.Project.Name)
		require.Equal(t, createInfo.Description, project.ProjectInfo.Description)
		require.Equal(t, createInfo.Metadata, project.ProjectInfo.Metadata)
		// validate error for attempting to get non-existent project.
		_, err = pfsdb.GetProjectInfo(cbCtx, tx, 3)
		require.YesError(t, err, "should not be able to get non-existent project")
		fmt.Println(err)
		require.True(t, (&pfsdb.ProjectNotFoundError{ID: 3}).Is(err))
		return nil
	}))
}

func TestForEachProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.DeleteProject(cbCtx, tx, "default"))
		size := 210
		expectedInfos := make([]*pfs.ProjectInfo, size)
		for i := 0; i < size; i++ {
			createInfo := &pfs.ProjectInfo{
				Project: &pfs.Project{
					Name: fmt.Sprintf("%s%d", testProj, i),
				},
				Description: testProjDesc,
				Metadata:    testMetadata,
				CreatedAt:   timestamppb.Now(),
			}
			expectedInfos[i] = createInfo
			require.NoError(t, pfsdb.CreateProject(ctx, tx, createInfo), "should be able to create project")
		}
		i := 0
		require.NoError(t, pfsdb.ForEachProject(cbCtx, tx, func(proj pfsdb.Project) error {
			require.Equal(t, expectedInfos[i].Project.Name, proj.ProjectInfo.Project.Name)
			require.Equal(t, expectedInfos[i].Description, proj.ProjectInfo.Description)
			require.Equal(t, expectedInfos[i].Metadata, proj.ProjectInfo.Metadata)
			i++
			return nil
		}))
		return nil
	}))
}

func TestUpdateProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// test upsert correctness
		require.YesError(t, pfsdb.UpdateProject(cbCtx, tx, 99, projInfo), "should not be able to create project when id is out of bounds")
		require.NoError(t, pfsdb.UpsertProject(cbCtx, tx, projInfo), "should be able to create project via upsert")
		projInfo.Description = "new desc"
		require.NoError(t, pfsdb.UpsertProject(cbCtx, tx, projInfo), "should be able to update project via upsert")
		projInfo.Metadata = testMetadata
		require.NoError(t, pfsdb.UpsertProject(cbCtx, tx, projInfo))
		return nil
	}))
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// test that after upserting the value actually changed in the database
		got, err := pfsdb.GetProjectInfoByName(cbCtx, tx, testProj)
		require.NoError(t, err, "should be able to get project %s", testProj)
		got.CreatedAt = nil
		if diff := cmp.Diff(projInfo, got, protocmp.Transform()); diff != "" {
			return errors.Errorf("GetProjectInfoByName(%s) (-want +got):\n%s", testProj, diff)
		}
		return nil
	}))
}

func TestUpdateProjectByID(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, pfsdb.CreateProject(cbCtx, tx, projInfo), "should be able to create project")
		// the 'default' project ID is 1
		projInfo.Metadata = testMetadata
		require.NoError(t, pfsdb.UpdateProject(cbCtx, tx, 2, projInfo), "should be able to update project")
		return nil
	}))
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// test that after upserting the value actually changed in the database
		got, err := pfsdb.GetProjectInfo(cbCtx, tx, 2)
		require.NoError(t, err, "should be able to get project 2")
		got.CreatedAt = nil
		if diff := cmp.Diff(projInfo, got, protocmp.Transform()); diff != "" {
			return errors.Errorf("GetProjectInfo(2) (-want +got):\n%s", diff)
		}
		return nil
	}))
}

func testProjectPicker() *pfs.ProjectPicker {
	return &pfs.ProjectPicker{
		Picker: &pfs.ProjectPicker_Name{
			Name: "default",
		},
	}
}

func TestPickProject(t *testing.T) {
	t.Parallel()
	namePicker := testProjectPicker()
	badProjectPicker := proto.Clone(namePicker).(*pfs.ProjectPicker)
	badProjectPicker.Picker.(*pfs.ProjectPicker_Name).Name = "does not exist"
	expected := &pfsdb.Project{ID: 1}
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		// default project already exists after migration, so no need to create it.
		got, err := pfsdb.PickProject(ctx, namePicker, tx)
		require.NoError(t, err, "should be able to pick project")
		require.Equal(t, expected.ID, got.ID)
		_, err = pfsdb.PickProject(ctx, nil, tx)
		require.YesError(t, err, "should error with a nil picker")
		_, err = pfsdb.PickProject(ctx, badProjectPicker, tx)
		require.YesError(t, err, "pick project should error with bad picker")
		require.True(t, errors.As(err, &pfsdb.ProjectNotFoundError{}))
	})
}
