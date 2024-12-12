//go:build k8s

package main

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"strconv"
	"testing"
)

type upgradeFunc func(t *testing.T, ctx context.Context, c *client.APIClient, fromVersion string)

// helmValuesPreGoCDK returns two maps. The first is for overriding SetValues, the second is for
// overriding SetStrValues.
func helmValuesPreGoCDK(numPachds int) (map[string]string, map[string]string) {
	return map[string]string{
			"pachw.minReplicas": "1",
			"pachw.maxReplicas": "5",
			"pachd.replicas":    strconv.Itoa(numPachds),
			// We are using "old" minio values here to pass CI tests. Current configurations has enabled gocdk by default,
			// so to make UpgradeTest work, we overried configuration with these "old" minio values.
			"pachd.storage.gocdkEnabled":   "false",
			"pachd.storage.backend":        "MINIO",
			"pachd.storage.minio.bucket":   "pachyderm-test",
			"pachd.storage.minio.endpoint": "minio.default.svc.cluster.local:9000",
			"pachd.storage.minio.id":       "minioadmin",
			"pachd.storage.minio.secret":   "minioadmin",
		},
		map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "false",
		}
}
