package coredb

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"testing"
)

const (
	testProj     = "testproj"
	testProjDesc = "this is a test project"
)

func TestCreateProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
	getInfo, err := GetProject(ctx, tx, testProj)
	require.NoError(t, err, "should be able to get a project")
	require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
	require.Equal(t, createInfo.Description, getInfo.Description)
}

func TestDeleteProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
	require.NoError(t, DeleteProject(ctx, tx, createInfo.Project.Name), "should be able to delete project")
	_, err = GetProject(ctx, tx, testProj)
	require.YesError(t, err, "get project should not find row")

}

func TestGetProjectByID(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
	require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
	getInfo, err := GetProjectByID(ctx, tx, 1)
	require.NoError(t, err, "should be able to get a project")
	require.NoError(t, tx.Commit())
	require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
	require.Equal(t, createInfo.Description, getInfo.Description)
}

func TestUpdateProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	// test upsert correctness
	projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
	require.YesError(t, UpdateProject(ctx, tx, projInfo, false), "should not be able to create project when upsert = false")
	require.NoError(t, UpdateProject(ctx, tx, projInfo, true), "should be able to create project when upsert = true")
	projInfo.Description = "new desc"
	require.NoError(t, UpdateProject(ctx, tx, projInfo, false), "should be able to update project")
}

func TestUpdateProjectByID(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
	require.NoError(t, CreateProject(ctx, tx, projInfo), "should be able to create project")
	require.NoError(t, UpdateProjectByID(ctx, tx, 1, projInfo), "should be able to update project")
}
