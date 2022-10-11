package testutil

import (
	"os"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

// GetTestEnterpriseCode Pulls the enterprise code out of the env var stored in CI
func GetTestEnterpriseCode(t testing.TB) string {
	acode, exists := os.LookupEnv("ENT_ACT_CODE")
	if !exists {
		t.Error("Enterprise activation code not found in environment variable (ENT_ACT_CODE)")
	}
	return acode
}

// ActivateEnterprise activates enterprise in Pachyderm (if it's not on already.)
func ActivateEnterprise(t testing.TB, c *client.APIClient, port ...string) {
	licensePort := "1650"
	if len(port) != 0 {
		licensePort = port[0]
	}
	ActivateLicense(t, c, licensePort)
	_, err := c.Enterprise.Activate(c.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "grpc://localhost:" + licensePort,
		})
	require.NoError(t, err)
}

func ActivateLicense(t testing.TB, c *client.APIClient, port string, expireTime ...time.Time) {
	code := GetTestEnterpriseCode(t)
	activateReq := &license.ActivateRequest{
		ActivationCode: code,
	}
	if len(expireTime) != 0 {
		activateReq.Expires = TSProtoOrDie(t, expireTime[0])
	}
	_, err := c.License.Activate(c.Ctx(), activateReq)
	require.NoError(t, err)
	_, err = c.License.AddCluster(c.Ctx(),
		&license.AddClusterRequest{
			Id:               "localhost",
			Secret:           "localhost",
			Address:          "grpc://localhost:" + port,
			UserAddress:      "grpc://localhost:" + port,
			EnterpriseServer: true,
		})
	if err != nil && !license.IsErrDuplicateClusterID(err) {
		require.NoError(t, err)
	}
}
