package coredb

import (
	"github.com/gogo/protobuf/types"
	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"testing"
	"time"
)

const (
	testProj     = "testproj"
	testProjDesc = "this is a test project"
)

func TestProjectCreateAndGet(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	require.NoError(t, v2_7_0.SetupCore(ctx, tx), "should be able to set up core tables")
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc}
	require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
	getInfo, err := GetProject(ctx, tx, testProj)
	require.NoError(t, err, "should be able to get a project")
	require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
	require.Equal(t, createInfo.Description, getInfo.Description)
	// test created_at and updated_at correctness
	createInfo.CreatedAt = types.TimestampNow()
	createInfo.Project.Name = "newproject"
	updateTime := types.TimestampNow()
	require.NoError(t, CreateProject(ctx, tx, createInfo, updateTime), "should be able to create project and override timestamps")
	rows := tx.QueryRowContext(ctx, "SELECT created_at, updated_at FROM core.projects WHERE name = 'newproject'")
	var createdAt, updatedAt time.Time
	require.NoError(t, rows.Scan(&createdAt, &updatedAt), "should be able to get timestamps")
	createdAtProto, err := types.TimestampProto(createdAt)
	require.NoError(t, err, "should be able to convert createdAt to proto timestamp")
	updatedAtProto, err := types.TimestampProto(updatedAt)
	require.NoError(t, err, "should be able to convert updatedAt to proto timestamp")
	require.Equal(t, createInfo.CreatedAt.Seconds, createdAtProto.Seconds)
	require.Equal(t, updateTime.Seconds, updatedAtProto.Seconds)
}

func TestGetProjectByID(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	require.NoError(t, v2_7_0.SetupCore(ctx, tx), "should be able to set up core tables")
	createInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
	require.NoError(t, CreateProject(ctx, tx, createInfo), "should be able to create project")
	getInfo, err := GetProjectByID(ctx, tx, 1)
	require.NoError(t, err, "should be able to get a project")
	require.NoError(t, tx.Commit())
	require.Equal(t, createInfo.Project.Name, getInfo.Project.Name)
	require.Equal(t, createInfo.Description, getInfo.Description)
	require.Equal(t, createInfo.CreatedAt.Seconds, getInfo.CreatedAt.Seconds)
}

func TestUpdateProject(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	require.NoError(t, v2_7_0.SetupCore(ctx, tx), "should be able to set up core tables")
	// test upsert correctness
	projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
	require.YesError(t, UpdateProject(ctx, tx, projInfo, false), "should not be able to create project when upsert = false")
	require.NoError(t, UpdateProject(ctx, tx, projInfo, true), "should be able to create project when upsert = true")
	projInfo.Description = "new desc"
	require.NoError(t, UpdateProject(ctx, tx, projInfo, false), "should be able to update project")
	// test update time correctness
	currTimeProto := types.TimestampNow()
	require.NoError(t, UpdateProject(ctx, tx, projInfo, false, currTimeProto), "should be able to update project with new time")
	row := tx.QueryRowContext(ctx, "SELECT updated_at FROM core.projects WHERE name = $1", projInfo.Project.Name)
	var dbTime time.Time
	require.NoError(t, row.Scan(&dbTime), "should be able to get database updated_at time")
	dbTimeProto, err := types.TimestampProto(dbTime)
	require.NoError(t, err, "should be able to convert time.Time to types.Timestamp")
	require.Equal(t, dbTimeProto.Seconds, currTimeProto.Seconds)
}

func TestUpdateProjectByID(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err, "should be able to create a tx")
	defer tx.Rollback()
	require.NoError(t, v2_7_0.SetupCore(ctx, tx), "should be able to set up core tables")
	projInfo := &pfs.ProjectInfo{Project: &pfs.Project{Name: testProj}, Description: testProjDesc, CreatedAt: types.TimestampNow()}
	require.NoError(t, CreateProject(ctx, tx, projInfo), "should be able to create project")
	require.NoError(t, UpdateProjectByID(ctx, tx, 1, projInfo), "should be able to update project")
}
