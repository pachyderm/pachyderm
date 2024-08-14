//go:build k8s

package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestImportKube(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.NoError(t, run(t, `
		pachctl config import-kube imported
		pachctl config get active-context | match 'imported'
		pachctl config get context imported | match "\"cluster_name\":[[:space:]]*\"$(kubectl config current-context)\""
		pachctl config get context imported | match '"namespace":[[:space:]]*"default"'
		pachctl config import-kube enterprise-kube --overwrite --namespace enterprise --enterprise
		pachctl config get active-enterprise-context | match 'enterprise-kube'
		pachctl config get context enterprise-kube | match "\"cluster_name\":[[:space:]]*\"$(kubectl config current-context)\""
		pachctl config get context enterprise-kube | match '"namespace":[[:space:]]*"enterprise"'

	`))
}
