package cmds

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/fsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
)

const (
	branch = "branch"
	commit = "commit"
	file   = "file"
	repo   = "repo"

	copy      = "copy"
	create    = "create"
	delete    = "delete"
	diff      = "diff"
	finish    = "finish"
	get       = "get"
	glob      = "glob"
	inspect   = "inspect"
	list      = "list"
	put       = "put"
	squash    = "squash"
	start     = "start"
	subscribe = "subscribe"
	update    = "update"
	wait      = "wait"
)

func mockInspectCluster(env *realenv.RealEnv) {
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		clusterInfo := admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}
		return &clusterInfo, nil
	})
}

func TestCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
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

		# Create a new project
		pachctl create project {{.project}}

		# Create a new repo with the same name
		pachctl create repo {{.repo}} --project {{.project}}

		# Create a commit and put some data in it
		commit3=$(pachctl start commit {{.repo}}@master --project {{.project}})
		echo "file contents" | pachctl put file {{.repo}}@${commit3}:/file -f - --project {{.project}}
		pachctl finish commit {{.repo}}@${commit3} --project {{.project}}

		# Set the current project
		pachctl config update context --project {{.project}}

		# The new commit should appear when using the new project
		pachctl list commit | match ${commit3}
		`,
		"repo", tu.UniqueString("TestCommit-repo"),
		"project", tu.UniqueString("project"),
	).Run())
}

func TestPutFileFullPathNoFilePath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create a temporary directory and file with nested structure
	tmpDir, err := os.MkdirTemp("", "pachyderm_test_put_file_full_path")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	nestedDir := filepath.Join(tmpDir, "nested/dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	filePath := filepath.Join(nestedDir, "testfile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test data"), 0644))

	repoName := tu.UniqueString("TestPutFileFullPathNoFilePath")

	// Create repo and put file with fullPath flag, and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master -f {{.filePath}} --full-path
        pachctl get file "{{.repo}}@master:{{.filePath}}" \
          | match "test data"
    `, "repo", repoName, "filePath", filePath).Run())
}

func TestPutFileFullPathWithFilePath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create a temporary file in a nested directory
	tmpDir, err := os.MkdirTemp("", "pachyderm_test_full_path_with_file_path")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	nestedDir := filepath.Join(tmpDir, "nested/dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	filePath := filepath.Join(nestedDir, "testfile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test data"), 0644))

	repoName := tu.UniqueString("TestPutFileFullPathWithFilePath")
	targetFilePath := "specific/target/file.txt"

	// Create repo, put file with file.Path and fullPath flag, and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master:{{.targetFilePath}} -f {{.filePath}} --full-path
        pachctl get file "{{.repo}}@master:{{.targetFilePath}}" \
          | match "test data"
    `, "repo", repoName, "targetFilePath", targetFilePath, "filePath", filePath).Run())
}

func TestPutFileFullPathWithFilePathEndingSlash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create a temporary file in a nested directory
	tmpDir, err := os.MkdirTemp("", "pachyderm_test_full_path_with_file_path_ending_slash")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	nestedDir := filepath.Join(tmpDir, "nested/dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	filePath := filepath.Join(nestedDir, "testfile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test data"), 0644))

	repoName := tu.UniqueString("TestPutFileFullPathWithFilePathEndingSlash")
	targetPrefix := "nested/dir/"

	// Create repo, put file with file.Path ending with '/' and fullPath flag, and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master:{{.targetPrefix}} -f {{.filePath}} --full-path
        pachctl get file "{{.repo}}@master:{{.targetPrefix}}{{.filePath}}" \
          | match "test data"
    `, "repo", repoName, "targetPrefix", targetPrefix, "filePath", filePath).Run())
}

func TestPutFileFilePathEndsWithSlashSingleSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "pachyderm_test_put_file_single_source")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) //nolint:errcheck

	require.NoError(t, os.WriteFile(tmpFile.Name(), []byte("test data"), 0644))

	repoName := tu.UniqueString("TestPutFileFilePathEndsWithSlashSingleSource")
	targetPrefix := "dir/"

	// Create repo, put file with file.Path ending with '/', and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master:{{.targetPrefix}} -f {{.fileName}}
        pachctl get file "{{.repo}}@master:{{.targetPrefix}}{{.baseFileName}}" \
          | match "test data"
    `, "repo", repoName, "targetPrefix", targetPrefix, "fileName", tmpFile.Name(), "baseFileName", filepath.Base(tmpFile.Name())).Run())
}

func TestPutFileFilePathWithoutSlashSingleSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "pachyderm_test_put_file_single_source_no_slash")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) //nolint:errcheck

	require.NoError(t, os.WriteFile(tmpFile.Name(), []byte("test data"), 0644))

	repoName := tu.UniqueString("TestPutFileFilePathWithoutSlashSingleSource")
	targetFilePath := "specific/target/path/file.txt"

	// Create repo, put file with file.Path without ending '/', and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master:{{.targetFilePath}} -f {{.fileName}}
        pachctl get file "{{.repo}}@master:{{.targetFilePath}}" \
          | match "test data"
    `, "repo", repoName, "targetFilePath", targetFilePath, "fileName", tmpFile.Name()).Run())
}

func TestPutFileRecursiveFullPathWithFilePathOfSlash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create two temporary files in a nested directory
	tmpDir, err := os.MkdirTemp("", "pachyderm_test_full_path_recursive_with_file_path_of_slash")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	nestedDir := filepath.Join(tmpDir, "nested/dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	filePath1 := filepath.Join(nestedDir, "testfile1.txt")
	require.NoError(t, os.WriteFile(filePath1, []byte("test data 1"), 0644))

	filePath2 := filepath.Join(nestedDir, "testfile2.txt")
	require.NoError(t, os.WriteFile(filePath2, []byte("test data 2"), 0644))

	repoName := tu.UniqueString("TestPutFileRecursiveFullPathWithFilePathOfSlash")
	targetPrefix := "/"

	// Create repo, put file with recursive and fullPath flag, and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master:{{.targetPrefix}} -r -f {{.nestedDir}} --full-path
        pachctl get file "{{.repo}}@master:{{.filePath1}}" \
          | match "test data 1"
		pachctl get file "{{.repo}}@master:{{.filePath2}}" \
          | match "test data 2"
    `, "repo", repoName, "targetPrefix", targetPrefix, "nestedDir", nestedDir, "filePath1", filePath1, "filePath2", filePath2).Run())
}

func TestPutFileRecursiveWithFilePathOfSlash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Create two temporary files in a nested directory
	tmpDir, err := os.MkdirTemp("", "pachyderm_test_recursive_with_file_path_of_slash")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	nestedDir := filepath.Join(tmpDir, "nested/dir")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	filePath1 := filepath.Join(nestedDir, "testfile1.txt")
	require.NoError(t, os.WriteFile(filePath1, []byte("test data 1"), 0644))

	filePath2 := filepath.Join(nestedDir, "testfile2.txt")
	require.NoError(t, os.WriteFile(filePath2, []byte("test data 2"), 0644))

	repoName := tu.UniqueString("TestPutFileRecursiveWithFilePathOfSlash")
	targetPrefix := "/"

	// Create repo, put file with recursive and fullPath flag, and verify
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
        pachctl create repo {{.repo}}
        pachctl put file {{.repo}}@master:{{.targetPrefix}} -r -f {{.nestedDir}}
        pachctl get file "{{.repo}}@master:{{.filePath1}}" \
          | match "test data 1"
		pachctl get file "{{.repo}}@master:{{.filePath2}}" \
          | match "test data 2"
    `, "repo", repoName, "targetPrefix", targetPrefix, "nestedDir", nestedDir, "filePath1", filepath.Base(filePath1), "filePath2", filepath.Base(filePath2)).Run())
}

func TestPutFileTAR(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// Test .tar file.
	require.NoError(t, fsutil.WithTmpFile("pachyderm_test_put_file_tar", func(f *os.File) error {
		require.NoError(t, tarutil.WithWriter(f, func(tw *tar.Writer) error {
			for i := 0; i < 3; i++ {
				name := strconv.Itoa(i)
				require.NoError(t, tarutil.WriteFile(tw, tarutil.NewMemFile(name, []byte(name))))
			}
			return nil
		}))

		name := f.Name()
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo {{.repo}}
		pachctl put file {{.repo}}@master -f {{.name}} --untar
		pachctl get file "{{.repo}}@master:/0" \
		  | match "0"
		pachctl get file "{{.repo}}@master:/1" \
		  | match "1"
		pachctl get file "{{.repo}}@master:/2" \
		  | match "2"
		`, "repo", tu.UniqueString("TestPutFileTAR"), "name", name).Run())
		return nil
	}, "tar"))
	// Test .tar.gz. file.
	require.NoError(t, fsutil.WithTmpFile("pachyderm_test_put_file_tar", func(f *os.File) error {
		gw := gzip.NewWriter(f)
		require.NoError(t, tarutil.WithWriter(gw, func(tw *tar.Writer) error {
			for i := 0; i < 3; i++ {
				name := strconv.Itoa(i)
				require.NoError(t, tarutil.WriteFile(tw, tarutil.NewMemFile(name, []byte(name))))
			}
			return nil
		}))
		require.NoError(t, gw.Close())
		name := f.Name()
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo {{.repo}}
		pachctl put file {{.repo}}@master -f {{.name}} --untar
		pachctl get file "{{.repo}}@master:/0" \
		  | match "0"
		pachctl get file "{{.repo}}@master:/1" \
		  | match "1"
		pachctl get file "{{.repo}}@master:/2" \
		  | match "2"
		`, "repo", tu.UniqueString("TestPutFileTAR"), "name", name).Run())
		return nil
	}, "tar", "gz"))
}

func TestPutFileNonexistentRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	repoName := tu.UniqueString("TestPutFileNonexistentRepo-repo")
	// This assumes that the file-existence check is after the
	// repo-existence check.  If you are seeing this test fail after
	// restructuring `pachctl put file`, then that is probably why; either
	// restore the order or adopt a different test.
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
                (pachctl put file {{.repo}}@master:random -f nonexistent-file 2>&1 || true) \
                  | match "repo default/{{.repo}} not found"
`,
		"repo", repoName).Run())
}

func TestMountParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	expected := map[string]*fuse.RepoOptions{
		"repo1": {
			Name:  "repo1", // name of mount, i.e. where to mount it
			File:  client.NewFile(pfs.DefaultProjectName, "repo1", "branch", "", ""),
			Write: true,
		},
		"repo2": {
			Name:  "repo2",
			File:  client.NewFile(pfs.DefaultProjectName, "repo2", "master", "", ""),
			Write: true,
		},
		"repo3": {
			Name: "repo3",
			File: client.NewFile(pfs.DefaultProjectName, "repo3", "master", "", ""),
		},
		"repo4": {
			Name: "repo4",
			File: client.NewFile(pfs.DefaultProjectName, "repo4", "master", "dee0c3904d6f44beb4fa10fc0db12d02", ""),
		},
	}
	opts, err := parseRepoOpts(pfs.DefaultProjectName, []string{"repo1@branch+w", "repo2+w", "repo3", "repo4@master=dee0c3904d6f44beb4fa10fc0db12d02"})
	require.NoError(t, err)
	for repo, ro := range expected {
		require.Equal(t, ro, opts[repo])
	}
}

func TestDiffFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project {{.project}}
		pachctl create repo {{.repo}} --project {{.project}}

		echo "foo" | pachctl put file {{.repo}}@master:/data --project {{.project}}

		echo "bar" | pachctl put file {{.repo}}@master:/data --project {{.project}}

		pachctl diff file {{.repo}}@master:/data {{.repo}}@master^:/data --project {{.project}} \
			| match -- '-foo'

		pachctl create project {{.otherProject}}
		pachctl create repo {{.repo}} --project {{.otherProject}}

		echo "foo" | pachctl put file {{.repo}}@master:/data --project {{.otherProject}}

		pachctl diff file {{.repo}}@master:/data {{.repo}}@master:/data --project {{.project}} \
			--old-project {{.otherProject}} | match -- '-foo'
		`,
		"repo", tu.UniqueString("TestDiffFile-repo"),
		"project", tu.UniqueString("TestDiffFile-project"),
		"otherProject", tu.UniqueString("TestDiffFile-project"),
	).Run())
}

func TestGetFileError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	repo := tu.UniqueString(t.Name())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo {{.repo}}

		pachctl put file {{.repo}}@master:/dir/foo <<EOF
		baz
		EOF

		pachctl put file {{.repo}}@master:/dir/bar <<EOF
		baz
		EOF
		`,
		"repo", repo,
	).Run())

	checkError := func(branch, path, message string) {
		req := fmt.Sprintf("%s@%s:%s", repo, branch, path)
		require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		(pachctl get file "{{.req}}" 2>&1 || true) | match "{{.message}}"
		`,
			"req", req,
			"message", message,
		).Run())
	}
	checkError("bad", "/foo", "not found")
	checkError("master", "/bad", "not found")
	checkError("master", "/dir", "Try again with the -r flag")
	checkError("master", "/dir/*", "Try again with the -r flag")
}

// TestSynonyms walks through the command tree for each resource and verb combination defined in PPS.
// A template is filled in that calls the help flag and the output is compared. It seems like 'match'
// is unable to compare the outputs correctly, but we can use diff here which returns an exit code of 0
// if there is no difference.
func TestSynonyms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	synonymCheckTemplate := `
		pachctl {{VERB}} {{RESOURCE_SYNONYM}} -h > synonym.txt
		pachctl {{VERB}} {{RESOURCE}} -h > singular.txt
		diff synonym.txt singular.txt
		rm synonym.txt singular.txt
	`

	resources := resourcesMap()
	synonyms := synonymsMap()

	for resource, verbs := range resources {
		withResource := strings.ReplaceAll(synonymCheckTemplate, "{{RESOURCE}}", resource)
		withResources := strings.ReplaceAll(withResource, "{{RESOURCE_SYNONYM}}", synonyms[resource])

		for _, verb := range verbs {
			synonymCommand := strings.ReplaceAll(withResources, "{{VERB}}", verb)
			t.Logf("Testing %s %s -h\n", verb, resource)
			require.NoError(t, testutilpachctl.BashCmd(synonymCommand).Run())
		}
	}
}

// TestSynonymsDocs is like TestSynonyms except it only tests commands registered by CreateDocsAliases.
func TestSynonymsDocs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	synonymCheckTemplate := `
		pachctl {{RESOURCE_SYNONYM}} -h > synonym.txt
		pachctl {{RESOURCE}} -h > singular.txt
		diff synonym.txt singular.txt
		rm synonym.txt singular.txt
	`

	synonyms := synonymsMap()

	for resource := range synonyms {
		withResource := strings.ReplaceAll(synonymCheckTemplate, "{{RESOURCE}}", resource)
		synonymCommand := strings.ReplaceAll(withResource, "{{RESOURCE_SYNONYM}}", synonyms[resource])

		t.Logf("Testing %s -h\n", resource)
		require.NoError(t, testutilpachctl.BashCmd(synonymCommand).Run())
	}
}

func TestProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	// env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	// c := env.PachClient
	// using xargs to trim newlines
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
                pachctl list project | xargs | match '^ACTIVE PROJECT CREATED DESCRIPTION \* default ([^-]+ ago) -$'
                pachctl create project foo
                pachctl list project | match "foo     ([^-]+ ago) -"
		`,
	).Run())
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
                pachctl create project foo
                `,
	).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
                pachctl update project foo -d "bar"
                pachctl inspect project foo | xargs | match "Name: foo Description: bar"
                pachctl delete project foo
                `,
	).Run())
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
                pachctl inspect project foo
                `,
	).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
                pachctl list project | xargs | match '^ACTIVE PROJECT CREATED DESCRIPTION \* default ([^-]+ ago) -$'
                pachctl create project foo
                `,
	).Run())
}

func resourcesMap() map[string][]string {
	return map[string][]string{
		branch: {create, delete, inspect, list},
		commit: {delete, finish, inspect, list, squash, start, subscribe, wait},
		file:   {copy, delete, diff, get, glob, inspect, list, put},
		repo:   {create, delete, inspect, list, update},
	}
}

func synonymsMap() map[string]string {
	return map[string]string{
		branch: branches,
		commit: commits,
		file:   files,
		repo:   repos,
	}
}

// TestMount creates two projects, each containing a repo with same name.  It
// mounts the repos and adds a single file to each, and verifies that the
// expected file appears in each.

func TestDeleteAllRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project {{.project}}
		pachctl create repo {{.repo}}
		pachctl create repo {{.repo}}a --project {{.project}}
		pachctl delete repo --all
		pachctl list repo --all-projects | match {{.repo}}a
		`,
		"project", tu.UniqueString("project"),
		"repo", tu.UniqueString("repo"),
	).Run())
}

func TestDeleteNonExistRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project {{.project}}
		pachctl delete repo {{.repo}},
		`,
		"project", tu.UniqueString("project"),
		"repo", tu.UniqueString("repo"),
	).Run())
}

func TestDeleteRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
			pachctl create project {{.project}}
			pachctl create repo {{.repo}} --project {{.project}}
			pachctl delete repo {{.repo}} --project {{.project}} 2>&1 | match 'Repo deleted.'
			`,
		"project", tu.UniqueString("project"),
		"repo", tu.UniqueString("repo"),
	).Run())
}

func TestDeleteAllReposAllProjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project {{.project}}
		pachctl create repo {{.repo}}
		pachctl create repo {{.repo}}a --project {{.project}}
		pachctl create project {{.project2}}
		pachctl create repo {{.repo}} --project {{.project2}}
		pachctl delete repo --all --all-projects
		if [ $(pachctl list repo --all-projects | wc -l) -ne 1 ]; then exit 1; fi
		`,
		"project", tu.UniqueString("project"),
		"project2", tu.UniqueString("project2"),
		"repo", tu.UniqueString("repo"),
	).Run())
}

func TestCmdListRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	project := tu.UniqueString("project")
	repo1, repo2 := tu.UniqueString("repo1"), tu.UniqueString("repo2")

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, "pachctl create project {{.project}}", "project", project).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, "pachctl create repo --project default {{.repo}}", "repo", repo1).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, "pachctl create repo --project {{.project}} {{.repo}}", "project", project, "repo", repo2).Run())

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl list repo | match {{.repo1}}
		pachctl list repo | match -v {{.repo2}}
		pachctl list repo --project {{.project}} | match {{.repo2}}
		pachctl list repo --all-projects | match {{.repo1}}
		pachctl list repo --all-projects | match {{.repo2}}
	`, "project", project, "repo1", repo1, "repo2", repo2).Run())
}

func TestBranchNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient

	project := tu.UniqueString("project")
	repo := tu.UniqueString("repo")

	if err := testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project {{.project}}
		pachctl create repo {{.repo}} --project {{.project}}
		(pachctl list file {{.repo}}@master --project {{.project}} 2>&1 && exit 1; true) | match 'branch "master" not found in repo {{.project}}/{{.repo}}'
		`, "project", project, "repo", repo).Run(); err != nil {
		t.Fatal(err)
	}
}

func TestBranchDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	pc := env.PachClient

	repo := tu.UniqueString("repo")
	if err := testutilpachctl.PachctlBashCmdCtx(ctx, t, pc, `
		pachctl create repo {{.repo}}
		pachctl create branch {{.repo}}@foo
	`, "repo", repo).Run(); err != nil {
		t.Fatal(err)
	}
	if err := testutilpachctl.PachctlBashCmdCtx(ctx, t, pc, `pachctl delete branch {{.repo}}@foo`, "repo", repo).Run(); err != nil {
		t.Fatal(err)
	}
	if err := testutilpachctl.PachctlBashCmdCtx(ctx, t, pc, `pachctl delete branch {{.repo}}@foo`, "repo", repo).Run(); err == nil {
		t.Error("want error for deleting a non-existent branch")
	}
}

func TestCopyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)

	// copy within default project
	srcRepo, destRepo := tu.UniqueString("srcRepo"), tu.UniqueString("destRepo")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		pachctl create repo {{.srcRepo}}
		pachctl create repo {{.destRepo}}
		echo "Lorem ipsum" | pachctl put file {{.srcRepo}}@master:/file

		pachctl copy file {{.srcRepo}}@master:/file {{.destRepo}}@master:/file
		pachctl get file {{.destRepo}}@master:/file | match "Lorem ipsum"
	`,
		"srcRepo", srcRepo,
		"destRepo", destRepo).Run())

	// copy within a user specified project
	project := tu.UniqueString("project")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		pachctl create project {{.project}}
		pachctl create repo --project {{.project}} {{.srcRepo}}
		pachctl create repo --project {{.project}} {{.destRepo}}
		echo "Lorem ipsum" | pachctl put file --project {{.project}} {{.srcRepo}}@master:/file

		pachctl copy file --project {{.project}} {{.srcRepo}}@master:/file {{.destRepo}}@master:/file
		pachctl get file --project {{.project}} {{.destRepo}}@master:/file | match "Lorem ipsum"
	`,
		"project", project,
		"srcRepo", srcRepo,
		"destRepo", destRepo).Run())

	// copy between projects
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		echo "Lorem ipsum" | pachctl put file {{.srcRepo}}@master:/file2
		echo "Lorem ipsum" | pachctl put file {{.srcRepo}}@master:/file3

		pachctl copy file --dest-project {{.project}} {{.srcRepo}}@master:/file2 {{.destRepo}}@master:/file2
		pachctl get file --project {{.project}} {{.destRepo}}@master:/file2 | match "Lorem ipsum"

		pachctl copy file --src-project default --dest-project {{.project}} {{.srcRepo}}@master:/file3 {{.destRepo}}@master:/file3
		pachctl get file --project {{.project}} {{.destRepo}}@master:/file3 | match "Lorem ipsum"
	`,
		"project", project,
		"srcRepo", srcRepo,
		"destRepo", destRepo).Run())
}

func TestDeleteProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	project := tu.UniqueString("project")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project {{.project}}
		pachctl create repo {{.repo}} --project {{.project}}
		pachctl put file --project {{.project}} {{.repo}}@master:/file <<<"This is a test"
		pachctl create pipeline --project {{.project}} <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
		        "repo": "{{.repo}}"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["cp /pfs/{{.repo}}/file /pfs/out"]
		    }
		  }
		EOF
		(pachctl delete project {{.project}} </dev/null) && exit 1
		(yes | pachctl delete project {{.project}}) || exit 1
		if [ $(pachctl list project | tail -n +2 | wc -l) -ne 1 ]; then exit 1; fi
		`,
		"project", project,
		"repo", tu.UniqueString("repo"),
		"pipeline", tu.UniqueString("pipeline"),
	).Run())
}

func TestInspectProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	mockInspectCluster(env)
	c := env.PachClient
	p, err := testutilpachctl.NewPachctl(ctx, c, fmt.Sprintf("%s/test-pach-config-%s.json", t.TempDir(), t.Name()))
	require.NoError(t, err, "NewPachctl must succeed")

	output, err := p.RunCommand(ctx, "pachctl inspect project default")
	require.NoError(t, err, "pachctl inspect project default must succeed:\n%s", output)
	require.True(t, strings.Contains(output, `Defaults: {}`), "pachctl inspect project must include defaults")

	output, err = p.RunCommand(ctx, `echo '{"createPipelineRequest": {"datumTries": 14}}' | pachctl update defaults --project default`)
	require.NoError(t, err, "pachctl update project defaults must succeed:\n%s", output)

	output, err = p.RunCommand(ctx, `pachctl inspect defaults --project default`)
	require.NoError(t, err, "pachctl inspect defaults must succeed:\n%s", output)

	output, err = p.RunCommand(ctx, "pachctl inspect project default")
	require.NoError(t, err, "pachctl inspect project default must succeed")
	require.True(t, strings.Contains(output, `Defaults: {"createPipelineRequest":{"datumTries":14}}`), "pachctl inspect project must include correct defaults; got %q", output)
}
