package cmds

import (
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func resourcesMap() map[string][]string {
	return map[string][]string{
		branch: {create, delete, inspect, list},
		commit: {delete, finish, inspect, list, squash, start, subscribe, wait},
		file:   {copy, delete, diff, get, glob, inspect, list, put},
		repo:   {create, delete, inspect, list, update},
	}
}

func pluralsMap() map[string]string {
	return map[string]string{
		branch: branches,
		commit: commits,
		file:   files,
		repo:   repos,
	}
}

func TestCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
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
	// TODO: Need to implement put file split in V2.
	t.Skip("Put file split not implemented in V2")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create repo {{.repo}}

		pachctl put file {{.repo}}@master:/data --split=csv --header-records=1 <<EOF
		name,job
		alice,accountant
		bob,baker
		EOF

		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match "name,job"
		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match "alice,accountant"
		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match -v "bob,baker"

		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match "name,job"
		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match "bob,baker"
		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match -v "alice,accountant"

		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "name,job"
		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "alice,accountant"
		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "bob,baker"
		`,
		"repo", tu.UniqueString("TestPutFileSplit-repo"),
	).Run())
}

func TestPutFileNonexistentRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	repoName := tu.UniqueString("TestPutFileNonexistentRepo-repo")
	// This assumes that the file-existence check is after the
	// repo-existence check.  If you are seeing this test fail after
	// restructuring `pachctl put file`, then that is probably why; either
	// restore the order or adopt a different test.
	require.NoError(t, tu.PachctlBashCmd(t, c, `
                (pachctl put file {{.repo}}@master:random -f nonexistent-file 2>&1 || true) \
                  | match "repo {{.repo}} not found"
`,
		"repo", repoName).Run())
}

func TestMountParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	expected := map[string]*fuse.RepoOptions{
		"repo1": {
			Name:   "repo1", // name of mount, i.e. where to mount it
			Repo:   "repo1",
			Branch: "branch",
			Write:  true,
		},
		"repo2": {
			Name:   "repo2",
			Repo:   "repo2",
			Branch: "master",
			Write:  true,
		},
		"repo3": {
			Name:   "repo3",
			Repo:   "repo3",
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

func TestDiffFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create repo {{.repo}}

		echo "foo" | pachctl put file {{.repo}}@master:/data

		echo "bar" | pachctl put file {{.repo}}@master:/data

		pachctl diff file {{.repo}}@master:/data {{.repo}}@master^:/data \
			| match -- '-foo'
		`,
		"repo", tu.UniqueString("TestDiffFile-repo"),
	).Run())
}

// TestPlurals walks through the command tree for each resource and verb combination defined in PPS.
// A template is filled in that calls the help flag and the output is compared. It seems like 'match'
// is unable to compare the outputs correctly, but we can use diff here which returns an exit code of 0
// if there is no difference.
func TestPlurals(t *testing.T) {
	pluralCheckTemplate := `
		pachctl {{VERB}} {{RESOURCE_PLURAL}} -h > plural.txt
		pachctl {{VERB}} {{RESOURCE}} -h > singular.txt
		diff plural.txt singular.txt
		rm plural.txt singular.txt
	`

	resources := resourcesMap()
	plurals := pluralsMap()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)

	for resource, verbs := range resources {
		withResource := strings.ReplaceAll(pluralCheckTemplate, "{{RESOURCE}}", resource)
		withResources := strings.ReplaceAll(withResource, "{{RESOURCE_PLURAL}}", plurals[resource])

		for _, verb := range verbs {
			pluralCommand := strings.ReplaceAll(withResources, "{{VERB}}", verb)
			t.Logf("Testing %s %s -h\n", verb, resource)
			require.NoError(t, tu.PachctlBashCmd(t, c, pluralCommand).Run())
		}
	}
}

// TestPluralsDocs is like TestPlurals except it only tests commands registered by CreateDocsAliases.
func TestPluralsDocs(t *testing.T) {
	pluralCheckTemplate := `
		pachctl {{RESOURCE_PLURAL}} -h > plural.txt
		pachctl {{RESOURCE}} -h > singular.txt
		diff plural.txt singular.txt
		rm plural.txt singular.txt
	`

	plurals := pluralsMap()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)

	for resource := range plurals {
		withResource := strings.ReplaceAll(pluralCheckTemplate, "{{RESOURCE}}", resource)
		pluralCommand := strings.ReplaceAll(withResource, "{{RESOURCE_PLURAL}}", plurals[resource])

		t.Logf("Testing %s -h\n", resource)
		require.NoError(t, tu.PachctlBashCmd(t, c, pluralCommand).Run())
	}
}
