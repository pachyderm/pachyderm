//go:build livek8s
// +build livek8s

package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestConfigListContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)

	// Verify the * marker exists for the active-context when enterprise is disabled and the enterprise context isn't set
	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo
		echo '{}' | pachctl config set context bar
		pachctl config set active-context bar
		pachctl config list context | match "\*	bar"
		pachctl config list context | match "	foo"
	`))

	require.NoError(t, tu.BashCmd(`
		echo {{.license}} | pachctl license activate
		pachctl enterprise get-state | match ACTIVE
		`,
		"license", tu.GetTestEnterpriseCode(t),
	).Run())

	require.NoError(t, run(t, `
		echo '{}' | pachctl config set context foo
		echo '{}' | pachctl config set context bar
		pachctl config set active-context bar
		pachctl config list context | match "E\*	bar"
		pachctl config list context | match "	foo"

		pachctl config set active-enterprise-context foo
		pachctl config list context | match "\*	bar"
		pachctl config list context | match "E	foo"

		pachctl config set active-context foo
		pachctl config list context | match "	bar"
		pachctl config list context | match "E\*	foo"
	`))
}

func TestImportKube(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	require.NoError(t, run(t, `
		pachctl config import-kube imported
		pachctl config get active-context | match 'imported'
		pachctl config get context imported | match '"cluster_name": "minikube"'
		pachctl config get context imported | match '"namespace": "default"'
		pachctl config import-kube enterprise-kube --overwrite --namespace enterprise --enterprise
		pachctl config get active-enterprise-context | match 'enterprise-kube'
		pachctl config get context enterprise-kube | match '"cluster_name": "minikube"'
		pachctl config get context enterprise-kube | match '"namespace": "enterprise"'

	`))
}
