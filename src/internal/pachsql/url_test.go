package pachsql

import (
	"testing"

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
