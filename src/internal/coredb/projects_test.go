//go:build unit_test

package coredb

import (
	"context"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	testProj     = "testproj"
	testProjDesc = "this is a test project"
)

func TestCreateProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, CreateProject(cbCtx, tx, createInfo), "should be able to create project")
		getInfo, err := GetProjectByName(cbCtx, tx, testProj)
		require.NoError(t, err, "should be able to get a project")
		require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		return nil
	}))
	require.YesError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		err := CreateProject(cbCtx, tx, createInfo)
		require.YesError(t, err, "should not be able to create project again with same name")
		require.True(t, ErrProjectAlreadyExists{testProj}.Is(err))
		fmt.Println("hello")
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
		createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
		require.NoError(t, CreateProject(cbCtx, tx, createInfo), "should be able to create project")
		require.NoError(t, DeleteProject(cbCtx, tx, createInfo.Project.Name), "should be able to delete project")
		_, err := GetProjectByName(cbCtx, tx, testProj)
		require.YesError(t, err, "get project should not find row")
		require.True(t, ErrProjectNotFound{Name: testProj}.Is(err))
		err = DeleteProject(cbCtx, tx, createInfo.Project.Name)
		require.YesError(t, err, "double delete should be an error")
		require.True(t, ErrProjectNotFound{Name: testProj}.Is(err))
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
		createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: timestamppb.Now()}
		require.NoError(t, CreateProject(cbCtx, tx, createInfo), "should be able to create project")
		// the 'default' project is ID 1.
		getInfo, err := GetProject(cbCtx, tx, 2)
		require.NoError(t, err, "should be able to get a project")
		require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
		require.Equal(t, createInfo.Description, getInfo.Description)
		_, err = GetProject(cbCtx, tx, 3)
		require.YesError(t, err, "should not be able to get non-existent project")
		require.True(t, ErrProjectNotFound{ID: 3}.Is(err))
		return nil
	}))
}

func TestListProject(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		require.NoError(t, DeleteProject(cbCtx, tx, "default"))
		size := 210
		expectedInfos := make([]*pfs.ProjectInfo, size)
		for i := 0; i < size; i++ {
			createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: fmt.Sprintf("%s%d", testProj, i)}, Description: testProjDesc, CreatedAt: timestamppb.Now()}
			expectedInfos[i] = createInfo
			require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
		}
		iter, err := ListProject(cbCtx, tx)
		require.NoError(t, err, "should be able to list projects")
		i := 0
		require.NoError(t, stream.ForEach[*pfs.ProjectInfo](cbCtx, iter, func(proj *pfs.ProjectInfo) error {
			if err != nil {
				require.NoError(t, err, "should be able to iterate over projects")
			}
			require.Equal(t, expectedInfos[i].Project.Name, proj.Project.Name)
			require.Equal(t, expectedInfos[i].Description, proj.Description)
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
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		// test upsert correctness
		projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: timestamppb.Now()}
		require.YesError(t, UpdateProject(cbCtx, tx, 99, projInfo), "should not be able to create project when id is out of bounds")
		require.NoError(t, UpsertProject(cbCtx, tx, projInfo), "should be able to create project via upsert")
		projInfo.Description = "new desc"
		require.NoError(t, UpsertProject(cbCtx, tx, projInfo), "should be able to update project via upsert")
		return nil
	}))
}

func TestUpdateProjectByID(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, dbutil.WithTx(ctx, db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: timestamppb.Now()}
		require.NoError(t, CreateProject(cbCtx, tx, projInfo), "should be able to create project")
		// the 'default' project ID is 1
		require.NoError(t, UpdateProject(cbCtx, tx, 2, projInfo), "should be able to update project")
		return nil
	}))
}
