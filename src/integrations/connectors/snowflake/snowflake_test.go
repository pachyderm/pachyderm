//go:build k8s

package snowflake

import (
	_ "embed"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
)

var (
	//go:embed snowflake-read.jsonnet
	snowflakeReadJsonnet string

	//go:embed snowflake-write.jsonnet
	snowflakeWriteJsonnet string
)

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

	fmt.Println(snowflakeReadJsonnet)
	fmt.Println(snowflakeWriteJsonnet)

	// - create input and output databases on Snowflake (ephemeral)
	// dbIn, dbInName := testsnowflake.NewEphemeralSnowflakeDB(t)
	// require.NoError(t, pachsql.CreateTestTable(db, "test_table", struct {
	// 	Id int    `column:"ID" dtype:"INT" constraint:"PRIMARY KEY"`
	// 	A  string `column:"A" dtype:"VARCHAR(100)"`
	// }{}))

	// - load some example data into input table
	// - create read pipeline that reads data from input table and writes to output repo
	// - create write pipeline that reads files from read's output repo and writes to output table
}
