package cmds

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}} --no-port-forwarding

		# Create a commit and put some data in it
		commit1=$(pachctl start commit {{.repo}}@master --no-port-forwarding)
		echo "file contents" | pachctl put file {{.repo}}@${commit1}:/file -f --no-port-forwarding -
		pachctl finish commit {{.repo}}@${commit1} --no-port-forwarding

		# Check that the commit above now appears in the output
		pachctl list commit {{.repo}} --no-port-forwarding \
		  | match ${commit1}

		# Create a second commit and put some data in it
		commit2=$(pachctl start commit {{.repo}}@master --no-port-forwarding)
		echo "file contents" | pachctl put file {{.repo}}@${commit2}:/file -f --no-port-forwarding -
		pachctl finish commit {{.repo}}@${commit2} --no-port-forwarding

		# Check that the commit above now appears in the output
		pachctl list commit {{.repo}} --no-port-forwarding \
		  | match ${commit1} \
		  | match ${commit2}
		`,
		"repo", tu.UniqueString("TestCommit-repo"),
	).Run())
}

func TestPutFileSplit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}} --no-port-forwarding

		pachctl put file {{.repo}}@master:/data --split=csv --header-records=1 --no-port-forwarding <<EOF
		name,job
		alice,accountant
		bob,baker
		EOF

		pachctl get file "{{.repo}}@master:/data/*0" --no-port-forwarding \
		  | match "name,job" \
			| match "alice,accountant" \
			| match -v "bob,baker"

		pachctl get file "{{.repo}}@master:/data/*1" --no-port-forwarding \
		  | match "name,job" \
			| match -v "alice,accountant" \
			| match "bob,baker"

		pachctl get file "{{.repo}}@master:/data/*" --no-port-forwarding \
		  | match "name,job" \
			| match "alice,accountant" \
			| match "bob,baker"
		`,
		"repo", tu.UniqueString("TestPutFileSplit-repo"),
	).Run())
}
