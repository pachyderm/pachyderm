package server

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	versionlib "github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"

	"github.com/golang/snappy"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KB = 1024
	MB = 1024 * KB
)

var pachClient *client.APIClient
var getPachClientOnce sync.Once

func getPachClient(t testing.TB) *client.APIClient {
	getPachClientOnce.Do(func() {
		var err error
		if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewOnUserMachine(false, false, "user")
		}
		require.NoError(t, err)
	})
	return pachClient
}

func collectCommitInfos(t testing.TB, commitInfoIter client.CommitInfoIterator) []*pfs.CommitInfo {
	var commitInfos []*pfs.CommitInfo
	for {
		commitInfo, err := commitInfoIter.Next()
		if err == io.EOF {
			return commitInfos
		}
		require.NoError(t, err)
		commitInfos = append(commitInfos, commitInfo)
	}
}

// TODO(msteffen) equivalent to funciton in src/server/auth/server/admin_test.go.
// These should be unified.
func RepoInfoToName(repoInfo interface{}) interface{} {
	return repoInfo.(*pfs.RepoInfo).Repo.Name
}

func TestExtractRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	nCommits := 2
	r := rand.New(rand.NewSource(45))
	fileContent := workload.RandString(r, 40*MB)
	for i := 0; i < nCommits; i++ {
		_, err := c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i), strings.NewReader(fileContent))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, "master"))
	}

	numPipelines := 3
	input := dataRepo
	for i := 0; i < numPipelines; i++ {
		pipeline := tu.UniqueString(fmt.Sprintf("TestExtractRestore%d", i))
		require.NoError(t, c.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
			},
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewPFSInput(input, "/*"),
			"",
			false,
		))
		input = pipeline
	}

	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	ops, err := c.ExtractAll(false) // TestExtractRestoreObjects tests 'true' case
	require.NoError(t, err)

	// Delete 1.7 data
	require.NoError(t, c.DeleteAll())

	// Restore metadata
	require.NoError(t, c.Restore(ops))

	commitIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitIter)
	require.Equal(t, numPipelines, len(commitInfos))

	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))
}

func TestExtractRestoreHeadlessBranches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	// create a headless branch
	require.NoError(t, c.CreateBranch(dataRepo, "headless", "", nil))

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.NoError(t, c.DeleteAll())
	require.NoError(t, c.Restore(ops))

	bis, err := c.ListBranch(dataRepo)
	require.NoError(t, err)
	require.Equal(t, 1, len(bis))
	require.Equal(t, "headless", bis[0].Branch.Name)
}

func TestExtractVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestore_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	r := rand.New(rand.NewSource(45))
	_, err := c.PutFile(dataRepo, "master", "file", strings.NewReader(workload.RandString(r, 40*MB)))
	require.NoError(t, err)

	pipeline := tu.UniqueString("TestExtractRestore")
	require.NoError(t, c.CreatePipeline(
		pipeline,
		"",
		[]string{"bash"},
		[]string{
			fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
		},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	ops, err := c.ExtractAll(false)
	require.NoError(t, err)
	require.True(t, len(ops) > 0)

	// Check that every Op looks right; the version set matches pachd's version
	for _, op := range ops {
		opV := reflect.ValueOf(op).Elem()
		var versions, nonemptyVersions int
		for i := 0; i < opV.NumField(); i++ {
			fDesc := opV.Type().Field(i)
			if !strings.HasPrefix(fDesc.Name, "Op") {
				continue
			}
			versions++
			if strings.HasSuffix(fDesc.Name,
				fmt.Sprintf("%d_%d", versionlib.MajorVersion, versionlib.MinorVersion)) {
				require.False(t, opV.Field(i).IsNil())
				nonemptyVersions++
			} else {
				require.True(t, opV.Field(i).IsNil())
			}
		}
		require.Equal(t, 1, nonemptyVersions)
		require.True(t, versions > 1)
	}
}

func TestMigrateFrom1_7(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Clear pachyderm cluster (so that next cluster starts up in a clean environment)
	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	// Restore dumped metadata (now that objects are present)
	md, err := os.Open(path.Join(os.Getenv("GOPATH"),
		"src/github.com/pachyderm/pachyderm/etc/testing/migration/1_7/sort.metadata"))
	require.NoError(t, err)
	require.NoError(t, c.RestoreReader(snappy.NewReader(md)))
	require.NoError(t, md.Close())

	// Wait for final imported commit to be processed
	commitIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit("left", "master")}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	// filter-left and filter-right both compute a join of left and
	// right--depending on when the final commit to 'left' was added, it may have
	// been processed multiple times (should be n * 3, as there are 3 pipelines)
	require.True(t, len(commitInfos) >= 3)

	// Inspect input
	commits, err := c.ListCommit("left", "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commits))
	commits, err = c.ListCommit("right", "master", "", 0)
	require.NoError(t, err)
	require.Equal(t, 3, len(commits))

	// Inspect output
	repos, err := c.ListRepo()
	require.NoError(t, err)
	require.ElementsEqualUnderFn(t,
		[]string{"left", "right", "copy", "sort"},
		repos, RepoInfoToName)

	// make sure all numbers 0-99 are in /nums
	var buf bytes.Buffer
	require.NoError(t, c.GetFile("sort", "master", "/nums", 0, 0, &buf))
	s := bufio.NewScanner(&buf)
	numbers := make(map[string]struct{})
	for s.Scan() {
		numbers[s.Text()] = struct{}{}
	}
	require.Equal(t, 100, len(numbers)) // job processed all inputs

	// Confirm stats commits are present
	commits, err = c.ListCommit("sort", "stats", "", 0)
	require.NoError(t, err)
	require.Equal(t, 6, len(commits))
}

func int64p(i int64) *int64 {
	return &i
}

// TestExtractRestoreObjects tests extraction and restoration of objects. Note
// that since 1.8, only data in input repos is referenced by objects, so this
// tests extracting/restoring an input repo.
func TestExtractRestoreObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := getPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestExtractRestoreObjects-input")
	require.NoError(t, c.CreateRepo(dataRepo))

	nCommits := 2
	r := rand.New(rand.NewSource(45))
	files := make(map[string]struct{})
	for i := 0; i < nCommits; i++ {
		hash := fnv.New64a()
		fileContent := workload.RandString(r, 40*MB)
		_, err := c.PutFile(dataRepo, "master", fmt.Sprintf("file-%d", i),
			io.TeeReader(strings.NewReader(fileContent), hash))
		require.NoError(t, err)
		files[string(hash.Sum(nil))] = struct{}{}
	}

	// Extract existing cluster state
	ops, err := c.ExtractAll(true)
	require.NoError(t, err)

	// Delete 1.7 data
	require.NoError(t, c.DeleteAll())
	// If test is running against minikube, delete old objects to test restoration
	// of objects
	kubectlContext, err := tu.Cmd("kubectl", "config", "current-context").Output()
	require.NoError(t, err)
	if strings.TrimSpace(string(kubectlContext)) == "minikube" {
		// 1. SSH into minikube
		minikubeIP, err := tu.Cmd("minikube", "ip").Output()
		minikubeIP = bytes.TrimSpace(minikubeIP)
		require.NoError(t, err)
		key, err := ioutil.ReadFile(path.Join(
			os.Getenv("HOME"), ".minikube/machines/minikube/id_rsa"))
		require.NoError(t, err)
		signer, err := ssh.ParsePrivateKey(key)
		require.NoError(t, err)
		config := &ssh.ClientConfig{
			User:            "docker",
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		sshClient, err := ssh.Dial("tcp", string(minikubeIP)+":22", config)
		require.NoError(t, err)

		// 2. remove old objects and blocks from minikube
		sesh, err := sshClient.NewSession()
		require.NoError(t, err)
		sesh.Stdout, sesh.Stderr = os.Stdout, os.Stderr
		require.NoError(t, sesh.Run("sudo rm -rf /var/pachyderm/pachd/pach/{object,block}/*"))

		// 3. Restart pachd pod (which likely has the data we just wrote in its
		// object cache
		kc := tu.GetKubeClient(t)
		pachdPods, err := kc.CoreV1().Pods("default").List(metav1.ListOptions{
			LabelSelector: "suite=pachyderm,app=pachd",
		})
		require.NoError(t, err)
		require.True(t, len(pachdPods.Items) > 0)
		for _, pod := range pachdPods.Items {
			kc.CoreV1().Pods("default").Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{
				GracePeriodSeconds: int64p(0),
			})
		}
		require.NoError(t, backoff.Retry(func() error {
			pachdPods, err = kc.CoreV1().Pods("default").List(metav1.ListOptions{
				LabelSelector: "suite=pachyderm,app=pachd",
			})
			require.NoError(t, err)
			if len(pachdPods.Items) == 0 {
				return fmt.Errorf("no restarted pachd pods are available yet")
			}
			var anyPodsRunning bool
			for _, pod := range pachdPods.Items {
				anyPodsRunning = anyPodsRunning || pod.Status.Phase == "Running"
			}
			if !anyPodsRunning {
				return fmt.Errorf("no restarted pachd pods are running yet")
			}
			return nil
		}, backoff.NewTestingBackOff()))

		// re-connect client & wait for response from pachd
		c = getPachClient(t)
		require.NoError(t, backoff.Retry(func() error {
			_, err := c.Version()
			return err
		}, backoff.NewTestingBackOff()))
	}

	// Restore metadata
	require.NoError(t, c.Restore(ops))

	// Check input data
	for i := 0; i < nCommits; i++ {
		hash := fnv.New64a()
		err := c.GetFile(dataRepo, "master", fmt.Sprintf("file-%d", i), 0, 0, hash)
		require.NoError(t, err)

		// Confirm that file contents didn't change
		hashval := string(hash.Sum(nil))
		_, ok := files[hashval]
		require.True(t, ok, "file %d had hash %q, which doesn't match what was written")

		// remove file, to check that one object wasn't deduped to the other
		delete(files, hashval)
	}
}
