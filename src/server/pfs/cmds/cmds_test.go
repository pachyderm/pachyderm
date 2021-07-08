package cmds_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/cmdtest"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/stretchr/testify/mock"
)

/*
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
	// TODO: Need to implement put file split in V2.
	t.Skip("Put file split not implemented in V2")
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

func TestMountParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
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

func TestDiffFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}}

		echo "foo" | pachctl put file {{.repo}}@master:/data

		echo "bar" | pachctl put file {{.repo}}@master:/data

		pachctl diff file {{.repo}}@master:/data {{.repo}}@master^:/data \
			| match -- '-foo'
		`,
		"repo", tu.UniqueString("TestDiffFile-repo"),
	).Run())
}
*/

func TestInspectRepo(suite *testing.T) {

	suite.Run("Success", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(cmdtest.RepoInfo(), nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo")
		require.NoError(t, err)
		require.Equal(t, "", stderr.String())
		require.Equal(t, 1, len(env.MockClient.PFS.Calls))
		require.Equal(t, &pfs.InspectRepoRequest{Repo: &pfs.Repo{Name: "foo", Type: pfs.UserRepoType}}, env.MockClient.PFS.Calls[0].Arguments[0])

		require.True(t, strings.Contains(stdout.String(), "foo"))
		require.True(t, strings.Contains(stdout.String(), "bar"))
		require.True(t, strings.Contains(stdout.String(), "100B"))
		require.True(t, strings.Contains(stdout.String(), "REPO_READ"))
		require.True(t, strings.Contains(stdout.String(), "REPO_WRITE"))
	})

	suite.Run("SpecRepo", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(cmdtest.RepoInfo(), nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo.spec")
		require.NoError(t, err)
		require.Equal(t, "", stderr.String())
		require.Equal(t, 1, len(env.MockClient.PFS.Calls))
		require.Equal(t, &pfs.InspectRepoRequest{Repo: &pfs.Repo{Name: "foo", Type: pfs.SpecRepoType}}, env.MockClient.PFS.Calls[0].Arguments[0])
	})

	suite.Run("UserRepo", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(cmdtest.RepoInfo(), nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo.user")
		require.NoError(t, err)
		require.Equal(t, "", stderr.String())
		require.Equal(t, 1, len(env.MockClient.PFS.Calls))
		require.Equal(t, &pfs.InspectRepoRequest{Repo: &pfs.Repo{Name: "foo", Type: pfs.UserRepoType}}, env.MockClient.PFS.Calls[0].Arguments[0])
	})

	suite.Run("ServerError", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		expectedErr := errors.New("fake server error")
		var nilRepoInfo *pfs.RepoInfo
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(nilRepoInfo, expectedErr)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo")
		require.YesError(t, err)
		require.Equal(t, expectedErr, err)
		require.Equal(t, "", stdout.String())
		require.Equal(t, "", stderr.String())
	})

	suite.Run("JSON", func(t *testing.T) {
		repoInfo := cmdtest.RepoInfo()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(repoInfo, nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo", "--raw", "--output=json")
		require.NoError(t, err)
		require.Equal(t, "", stderr.String())

		result := &pfs.RepoInfo{}
		require.NoError(t, serde.NewJSONDecoder(stdout).DecodeProto(result))
		require.Equal(t, repoInfo, result)
	})

	suite.Run("YAML", func(t *testing.T) {
		repoInfo := cmdtest.RepoInfo()
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(repoInfo, nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo", "--raw", "--output=yaml")
		require.NoError(t, err)
		require.Equal(t, "", stderr.String())

		result := &pfs.RepoInfo{}
		require.NoError(t, serde.NewYAMLDecoder(stdout).DecodeProto(result))
		require.Equal(t, repoInfo, result)
	})

	suite.Run("InvalidOutput", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(cmdtest.RepoInfo(), nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo", "--raw", "--output=hunter2")
		require.YesError(t, err)
		require.Equal(t, "", stdout.String())
		require.Equal(t, "", stderr.String())
		require.Matches(t, "unrecognized encoding", err.Error())
	})

	suite.Run("OutputWithoutRaw", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		env := cmdtest.NewTestEnv(t, &bytes.Buffer{}, stdout, stderr)
		env.MockClient.PFS.On("InspectRepo", mock.Anything).Return(cmdtest.RepoInfo(), nil)

		err := cmdtest.RunPachctlSync(t, env, "inspect", "repo", "foo", "--output=yaml")
		require.YesError(t, err)
		require.Equal(t, "", stdout.String())
		require.Equal(t, "", stderr.String())
		require.Matches(t, "cannot set --output", err.Error())
	})
