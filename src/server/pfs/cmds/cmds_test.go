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
		pachctl create-repo {{.repo}}

		# Create a commit and put some data in it
		commit1=$(pachctl start-commit {{.repo}} master)
		echo "file contents" | pachctl put-file {{.repo}} ${commit1} /file -f -
		pachctl finish-commit {{.repo}} ${commit1}

		# Check that the commit above now appears in the output
		pachctl list-commit {{.repo}} \
		  | match ${commit1}

		# Create a second commit and put some data in it
		commit2=$(pachctl start-commit {{.repo}} master)
		echo "file contents" | pachctl put-file {{.repo}} ${commit2} /file -f -
		pachctl finish-commit {{.repo}} ${commit2}

		# Check that the commit above now appears in the output
		pachctl list-commit {{.repo}} \
		  | match ${commit1} \
		  | match ${commit2}
		`,
		"repo", tu.UniqueString("TestCommit-repo"),
	).Run())
}
