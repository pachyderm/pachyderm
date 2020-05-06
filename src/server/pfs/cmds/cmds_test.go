package cmds

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}}

		# Create a commit and put some data in it
		commit1=$(pachctl start commit {{.repo}}@master)
		echo "file contents" | pachctl put file {{.repo}}@${commit1}:/file -f -
		pachctl finish commit {{.repo}}@${commit1}

		# Check that the commit above now appears in the output
		pachctl list commit {{.repo}} \
		  | match ${commit1}

		# Create a second commit and put some data in it
		commit2=$(pachctl start commit {{.repo}}@master)
		echo "file contents" | pachctl put file {{.repo}}@${commit2}:/file -f -
		pachctl finish commit {{.repo}}@${commit2}

		# Check that the commit above now appears in the output
		pachctl list commit {{.repo}} \
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
		pachctl create repo {{.repo}}

		pachctl put file {{.repo}}@master:/data --split=csv --header-records=1 <<EOF
		name,job
		alice,accountant
		bob,baker
		EOF

		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match "name,job" \
			| match "alice,accountant" \
			| match -v "bob,baker"

		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match "name,job" \
			| match -v "alice,accountant" \
			| match "bob,baker"

		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "name,job" \
			| match "alice,accountant" \
			| match "bob,baker"
		`,
		"repo", tu.UniqueString("TestPutFileSplit-repo"),
	).Run())
}

func TestMountParsing(t *testing.T) {
	expected := map[string]*fuse.RepoOptions{
		"repo1": {
			Branch: "branch",
			Write:  true,
		},
		"repo2": {
			Branch: "master",
			Write:  true,
		},
		"repo3": {
			Branch: "master",
		},
	}
	opts, err := parseRepoOpts([]string{"repo1@branch+w", "repo2+w", "repo3"})
	require.NoError(t, err)
	require.Equal(t, 3, len(opts))
	fmt.Printf("%+v\n", opts)
	for repo, ro := range expected {
		require.Equal(t, ro, opts[repo])
	}
}
