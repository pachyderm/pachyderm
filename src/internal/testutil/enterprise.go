package testutil

import (
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

// GetTestEnterpriseCode Pulls the enterprise code out of the env var stored in travis
func GetTestEnterpriseCode(t testing.TB) string {
	acode, exists := os.LookupEnv("ENT_ACT_CODE")
	if !exists {
		t.Error("Enterprise Activation code not found in Env Vars")
	}
	return acode

}

// ActivateEnterprise activates enterprise in Pachyderm (if it's not on already.)
func ActivateEnterprise(t testing.TB, c *client.APIClient) {
	code := GetTestEnterpriseCode(t)

	_, err := c.License.Activate(c.Ctx(),
		&license.ActivateRequest{
			ActivationCode: code,
		})
	require.NoError(t, err)

	_, err = c.License.AddCluster(c.Ctx(),
		&license.AddClusterRequest{
			Id:               "localhost",
			Secret:           "localhost",
			Address:          "grpc://localhost:650",
			UserAddress:      "grpc://localhost:650",
			EnterpriseServer: true,
		})
	if err != nil && !license.IsErrDuplicateClusterID(err) {
		require.NoError(t, err)
	}

	_, err = c.Enterprise.Activate(c.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "grpc://localhost:650",
		})
	require.NoError(t, err)
}
