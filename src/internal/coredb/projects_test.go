//go:build unit_test

package coredb

import (
	"fmt"
	"io"
	"testing"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	require.YesError(t, DeleteProject(ctx, tx, createInfo.Project.Name), "double delete should be an error")
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
	// the 'default' project is ID 1.
	getInfo, err := GetProjectByID(ctx, tx, 2)
	fmt.Printf("%q %q\n", createInfo, getInfo)
	require.NoError(t, err, "should be able to get a project")
	require.NoError(t, tx.Commit())
	require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
	require.Equal(t, createInfo.Description, getInfo.Description)
}

func TestListProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	require.NoError(t, DeleteProject(ctx, db, "default"))
	size := 210
	expectedInfos := make([]*pfs.ProjectInfo, size)
	for i := 0; i < size; i++ {
		createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: fmt.Sprintf("%s%d", testProj, i)}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
		expectedInfos[i] = createInfo
		require.NoError(t, CreateProject(ctx, db, createInfo), "should be able to create project")
	}
	iter, err := ListProject(ctx, db)
	require.NoError(t, err, "should be able to list projects")
	i := 0
	projectInfo := &pfs.ProjectInfo{}
	for err := iter.Next(ctx, projectInfo); !errors.Is(err, io.EOF); err = iter.Next(ctx, projectInfo) {
		if err != nil {
			require.NoError(t, err, "should be able to iterate over projects")
		}
		require.Equal(t, expectedInfos[i].Project.Name, projectInfo.Project.Name)
		require.Equal(t, expectedInfos[i].Description, projectInfo.Description)
		i++
	}
}

func TestDeleteAllProjects(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	size := 3
	for i := 0; i < size; i++ {
		createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: fmt.Sprintf("%s%d", testProj, i)}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
		require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
	}
	require.NoError(t, DeleteAllProjects(ctx, tx))
	_, err = GetProjectByID(ctx, tx, 1)
	require.YesError(t, err, "should not have any project entries")
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
	// the 'default' project ID is 1
	require.NoError(t, UpdateProjectByID(ctx, tx, 2, projInfo), "should be able to update project")
}
