package pachsql

import (
	"fmt"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	testCases := []struct {
		In  string
		Out URL
	}{
		{
			In: "postgres://10.0.0.1:9000/mydb?sslmode=disable",
			Out: URL{
				Protocol: "postgres",
				Host:     "10.0.0.1",
				Port:     9000,
				Database: "mydb",
				Params: map[string]string{
					"sslmode": "disable",
				},
			},
		},
		{
			In: "mysql://jbond@10.0.0.1:1007/martini?shaken=true&stirred=false",
			Out: URL{
				Protocol: "mysql",
				User:     "jbond",
				Host:     "10.0.0.1",
				Port:     1007,
				Database: "martini",
				Params: map[string]string{
					"shaken":  "true",
					"stirred": "false",
				},
			},
		},
		{
			In: "snowflake://jbond@mi6/martini?shaken=true&stirred=false",
			Out: URL{
				Protocol: "snowflake",
				User:     "jbond",
				Host:     "mi6",
				Port:     0,
				Database: "martini",
				Params: map[string]string{
					"shaken":  "true",
					"stirred": "false",
				},
			},
		},
		{
			In: "snowflake://jbond@mi6.snowflakecomputing.com:443/martini/schemaname?shaken=true&stirred=false",
			Out: URL{
				Protocol: "snowflake",
				User:     "jbond",
				Host:     "mi6.snowflakecomputing.com",
				Port:     443,
				Database: "martini",
				Schema:   "schemaname",
				Params: map[string]string{
					"shaken":  "true",
					"stirred": "false",
				},
			},
		},
	}
	for _, tc := range testCases {
		actual, err := ParseURL(tc.In)
		assert.NoError(t, err)
		assert.Equal(t, &tc.Out, actual)
	}
}

func TestSnowflakeConnection(t *testing.T) {
	user := os.Getenv("SNOWSQL_USER")
	role := os.Getenv("SNOWSQL_ROLE")
	account := os.Getenv("SNOWSQL_ACCOUNT")
	host := account + ".snowflakecomputing.com"
	password := os.Getenv("SNOWSQL_PWD")

	tests := map[string]struct {
		dsn     string
		require func(testing.TB, error, ...interface{})
	}{
		"simple":         {fmt.Sprintf("snowflake://%s@%s", user, account), require.NoError},
		"wh":             {fmt.Sprintf("snowflake://%s@%s?warehouse=COMPUTE_WH", user, account), require.NoError},
		"bad wh":         {fmt.Sprintf("snowflake://%s@%s?warehouse=badWH", user, account), require.YesError},
		"role":           {fmt.Sprintf("snowflake://%s@%s?role=%s", user, account, role), require.NoError},
		"role and my_wh": {fmt.Sprintf("snowflake://%s@%s?my_warehouse=COMPUTE_WH&role=%s", user, account, role), require.NoError},
		"role and wh":    {fmt.Sprintf("snowflake://%s@%s?warehouse=COMPUTE_WH&role=%s", user, account, role), require.NoError},
		"bad role":       {fmt.Sprintf("snowflake://%s@%s?role=badRole", user, account), require.YesError},
		"host":           {fmt.Sprintf("snowflake://%s@%s", user, host), require.NoError},
		"host with port": {fmt.Sprintf("snowflake://%s@%s:443?account=%s", user, host, account), require.NoError},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			url, err := ParseURL(tc.dsn)
			require.NoError(t, err)
			db, err := OpenURL(*url, password)
			require.NoError(t, err)
			tc.require(t, db.Ping())
		})
	}
}
