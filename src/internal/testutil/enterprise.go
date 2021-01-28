package testutil

import (
	"context"
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

	resp, err := c.Enterprise.GetState(context.Background(),
		&enterprise.GetStateRequest{})
	require.NoError(t, err)

	if resp.State == enterprise.State_ACTIVE {
		return
	}

	_, err = c.License.Activate(context.Background(),
		&license.ActivateRequest{
			ActivationCode: code,
		})
	require.NoError(t, err)

	_, err = c.License.AddCluster(context.Background(),
		&license.AddClusterRequest{
			Id:      "localhost",
			Secret:  "localhost",
			Address: "localhost:650",
		})
	if err != nil && license.IsErrDuplicateClusterID(err) {
		require.NoError(t, err)
	}

	_, err = c.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "localhost:650",
		})
	require.NoError(t, err)
}
