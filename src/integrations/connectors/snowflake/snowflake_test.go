//go:build k8s

package snowflake

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	sdataTU "github.com/pachyderm/pachyderm/v2/src/internal/sdata/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testsnowflake"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

var (
	//go:embed snowflake-read.jsonnet
	readTemplate string
	//go:embed snowflake-write.jsonnet
	writeTemplate string
)

type snowflakeRow struct {
	Id           int16         `column:"c_id" dtype:"SMALLINT" constraint:"PRIMARY KEY NOT NULL"`
	A            string        `column:"c_a" dtype:"VARCHAR(100)" constraint:"NOT NULL"`
	SmallintNull sql.NullInt16 `column:"c_smallint_null" dtype:"SMALLINT" constraint:"NULL"`
}

func (row *snowflakeRow) SetID(id int16) {
	row.Id = id
}

func TestSnowflake(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

	// create K8s secrets
	b := []byte(fmt.Sprintf(`
	{
		"apiVersion": "v1",
		"kind": "Secret",
		"stringData": {
			"SNOWSQL_PWD": "%s"
		},
		"metadata": {
			"name": "snowflake-secret",
			"creationTimestamp": null
		}
	}`, os.Getenv("SNOWFLAKE_PASSWORD")))
	require.NoError(t, c.CreateSecret(b))

	// create ephemeral input and output databases
	// testutil.Cleanup = false
	tableName := "test_table"
	inDB, inDBName := testsnowflake.NewEphemeralSnowflakeDB(t)
	require.NoError(t, pachsql.CreateTestTable(inDB, tableName, &snowflakeRow{}))
	outDB, outDBName := testsnowflake.NewEphemeralSnowflakeDB(t)
	require.NoError(t, pachsql.CreateTestTable(outDB, tableName, &snowflakeRow{}))

	// load some example data into input table
	nRows := 10
	require.NoError(t, sdataTU.GenerateTestData(inDB, tableName, nRows, &snowflakeRow{}))

	// create read pipeline that reads data from input table and writes to output repo
	ctx := context.Background()
	readPipeline, writePipeline := "read", "write"
	readPipelineTempl, err := c.RenderTemplate(ctx, &pps.RenderTemplateRequest{
		Args: map[string]string{
			// Pachyderm
			"image":    "pachyderm/snowflake:local",
			"name":     readPipeline,
			"cronSpec": "@yearly", // we want to manually trigger this
			// Snowflake
			"account":     os.Getenv("SNOWFLAKE_ACCOUNT"),
			"user":        os.Getenv("SNOWFLAKE_USER"),
			"role":        os.Getenv("SNOWFLAKE_USER_ROLE"),
			"warehouse":   "COMPUTE_WH",
			"database":    inDBName,
			"schema":      "public",
			"query":       fmt.Sprintf("select * from %s", tableName),
			"fileFormat":  `(type = csv FIELD_OPTIONALLY_ENCLOSED_BY = '0x22' COMPRESSION=NONE FILE_EXTENSION='')`,
			"outputFile":  "test_table/test_table.csv",
			"copyOptions": "OVERWRITE = TRUE SINGLE = TRUE",
		},
		Template: readTemplate,
	})
	require.NoError(t, err)
	writePipelineTempl, err := c.RenderTemplate(ctx, &pps.RenderTemplateRequest{
		Args: map[string]string{
			// Pachyderm
			"image":     "pachyderm/snowflake:local",
			"inputRepo": readPipeline,
			"name":      writePipeline,
			// Snowflake
			"account":    os.Getenv("SNOWFLAKE_ACCOUNT"),
			"user":       os.Getenv("SNOWFLAKE_USER"),
			"role":       os.Getenv("SNOWFLAKE_USER_ROLE"),
			"warehouse":  "COMPUTE_WH",
			"database":   outDBName,
			"schema":     "public",
			"table":      tableName,
			"fileFormat": `(type = csv FIELD_OPTIONALLY_ENCLOSED_BY = '0x22')`,
			// "copyOptions": "PURGE=TRUE",
		},
		Template: writeTemplate,
	})
	require.NoError(t, err)
	pipelineReader, err := ppsutil.NewPipelineManifestReader([]byte(fmt.Sprintf("[%s,%s]", readPipelineTempl.GetJson(), writePipelineTempl.GetJson())))
	require.NoError(t, err)
	for {
		request, err := pipelineReader.NextCreatePipelineRequest()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		_, err = c.PpsAPIClient.CreatePipeline(ctx, request)
		require.NoError(t, err)
	}

	// run cron job and wait for both pipelines to succeed
	require.NoError(t, c.RunCron(readPipeline))
	commitInfo, err := c.WaitCommit(readPipeline, "master", "")
	require.NoError(t, err)
	jobInfo, err := c.InspectJob(readPipeline, commitInfo.Commit.ID, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.GetState())

	files, err := c.ListFileAll(commitInfo.Commit, "/test_table")
	require.NoError(t, err)
	require.Len(t, files, 1)

	commitInfo, err = c.WaitCommit(writePipeline, "master", "")
	require.NoError(t, err)
	jobInfo, err = c.InspectJob(writePipeline, commitInfo.Commit.ID, false)
	require.NoError(t, err)
	require.Equal(t, pps.JobState_JOB_SUCCESS, jobInfo.GetState())

	// finally verify that the target table actually has any data
	var count int
	require.NoError(t, outDB.QueryRow("select count(*) from test_table").Scan(&count))
	require.Equal(t, nRows, count)
}
