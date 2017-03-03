package server

import (
	"fmt"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

//func TestPipelineWithFullObjects(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()
//c := getPachClient(t)
//// create repos
//dataRepo := uniqueString("TestPipeline_data")
//require.NoError(t, c.CreateRepo(dataRepo))
//// create pipeline
//pipelineName := uniqueString("pipeline")
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
//nil,
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{
//{
//Name:   dataRepo,
//Repo:   client.NewRepo(dataRepo),
//Glob:   "/*",
//Branch: "master",
//},
//},
//false,
//))
//// Do first commit to repo
//commit1, err := c.StartCommit(dataRepo, "master")
//require.NoError(t, err)
//_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
//require.NoError(t, err)
//require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
//commitInfos, err := c.FlushCommit([]*pfsclient.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
//require.NoError(t, err)
//require.Equal(t, 2, len(commitInfos))
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "foo\n", buffer.String())
//// Do second commit to repo
//commit2, err := c.StartCommit(dataRepo, commit1.ID)
//require.NoError(t, err)
//_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
//require.NoError(t, err)
//require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
//commitInfos, err = c.FlushCommit([]*pfsclient.Commit{client.NewCommit(dataRepo, commit2.ID)}, nil)
//require.NoError(t, err)
//require.Equal(t, 2, len(commitInfos))
//buffer = bytes.Buffer{}
//require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "foo\nbar\n", buffer.String())
//}

// TestChainedPipelines tracks https://github.com/pachyderm/pachyderm/issues/797
//func TestChainedPipelines(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()
//c := getPachClient(t)
//aRepo := uniqueString("A")
//require.NoError(t, c.CreateRepo(aRepo))

//dRepo := uniqueString("D")
//require.NoError(t, c.CreateRepo(dRepo))

//aCommit, err := c.StartCommit(aRepo, "master")
//require.NoError(t, err)
//_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
//require.NoError(t, err)
//require.NoError(t, c.FinishCommit(aRepo, "master"))

//dCommit, err := c.StartCommit(dRepo, "master")
//require.NoError(t, err)
//_, err = c.PutFile(dRepo, "master", "file", strings.NewReader("bar\n"))
//require.NoError(t, err)
//require.NoError(t, c.FinishCommit(dRepo, "master"))

//bPipeline := uniqueString("B")
//require.NoError(t, c.CreatePipeline(
//bPipeline,
//"",
//[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
//nil,
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{{Repo: client.NewRepo(aRepo)}},
//false,
//))

//cPipeline := uniqueString("C")
//require.NoError(t, c.CreatePipeline(
//cPipeline,
//"",
//[]string{"sh"},
//[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
//fmt.Sprintf("cp /pfs/%s/file /pfs/out/dFile", dRepo)},
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{{Repo: client.NewRepo(bPipeline)},
//{Repo: client.NewRepo(dRepo)}},
//false,
//))
//results, err := c.FlushCommit([]*pfsclient.Commit{aCommit, dCommit}, nil)
//require.NoError(t, err)
//require.Equal(t, 4, len(results))
//}

func TestChainedPipelinesNoDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
	aRepo := uniqueString("A")
	require.NoError(t, c.CreateRepo(aRepo))

	eRepo := uniqueString("E")
	require.NoError(t, c.CreateRepo(eRepo))

	aCommit, err := c.StartCommit(aRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(aRepo, aCommit.ID, "master"))
	_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, "master"))

	eCommit, err := c.StartCommit(eRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(eRepo, eCommit.ID, "master"))
	_, err = c.PutFile(eRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(eRepo, "master"))

	bPipeline := uniqueString("B")
	require.NoError(t, c.CreatePipeline(
		bPipeline,
		"",
		[]string{"cp", path.Join("/pfs", aRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name: aRepo,
			Repo: client.NewRepo(aRepo),
			Glob: "/",
		}},
		"",
		false,
	))

	cPipeline := uniqueString("C")
	require.NoError(t, c.CreatePipeline(
		cPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/file /pfs/out/bFile", bPipeline),
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/eFile", eRepo)},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name: bPipeline,
			Repo: client.NewRepo(bPipeline),
			Glob: "/",
		}, {
			Name: eRepo,
			Repo: client.NewRepo(eRepo),
			Glob: "/",
		}},
		"",
		false,
	))

	dPipeline := uniqueString("D")
	require.NoError(t, c.CreatePipeline(
		dPipeline,
		"",
		[]string{"sh"},
		[]string{fmt.Sprintf("cp /pfs/%s/bFile /pfs/out/bFile", cPipeline),
			fmt.Sprintf("cp /pfs/%s/eFile /pfs/out/eFile", cPipeline)},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name: cPipeline,
			Repo: client.NewRepo(cPipeline),
			Glob: "/",
		}},
		"",
		false,
	))

	resultsIter, err := c.FlushCommit([]*pfs.Commit{aCommit, eCommit}, nil)
	require.NoError(t, err)
	results, err := collectCommitInfos(resultsIter)
	require.NoError(t, err)
	require.Equal(t, 5, len(results))

	eCommit2, err := c.StartCommit(eRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(eRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(eRepo, "master"))

	resultsIter, err = c.FlushCommit([]*pfs.Commit{eCommit2}, nil)
	require.NoError(t, err)
	results, err = collectCommitInfos(resultsIter)
	require.NoError(t, err)
	require.Equal(t, 3, len(results))

	// Get number of jobs triggered in pipeline D
	jobInfos, err := c.ListJob(dPipeline, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
}

func collectCommitInfos(commitInfoIter client.CommitInfoIterator) ([]*pfs.CommitInfo, error) {
	var commitInfos []*pfs.CommitInfo
	for {
		commitInfo, err := commitInfoIter.Next()
		if err == io.EOF {
			return commitInfos, nil
		}
		if err != nil {
			return nil, err
		}
		commitInfos = append(commitInfos, commitInfo)
	}
}

func TestParallelismSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Test Constant strategy
	parellelism, err := pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy: pps.ParallelismSpec_CONSTANT,
		Constant: 7,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(7), parellelism)

	// Coefficient == 1 (basic test)
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy:    pps.ParallelismSpec_COEFFICIENT,
		Coefficient: 1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Coefficient > 1
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy:    pps.ParallelismSpec_COEFFICIENT,
		Coefficient: 2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2), parellelism)

	// Make sure we start at least one worker
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{
		Strategy:    pps.ParallelismSpec_COEFFICIENT,
		Coefficient: 0.1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Test 0-initialized JobSpec
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), &pps.ParallelismSpec{})
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)

	// Test nil JobSpec
	parellelism, err = pps_server.GetExpectedNumWorkers(getKubeClient(t), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(1), parellelism)
}

func TestPipelineJobDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"},
		nil,
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name: dataRepo,
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/",
		}},
		"",
		false,
	))

	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))

	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)

	_, err = commitIter.Next()
	require.NoError(t, err)

	// Now delete the corresponding job
	jobInfos, err := c.ListJob(pipelineName, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobInfos))
	err = c.DeleteJob(jobInfos[0].Job.ID)
	require.NoError(t, err)
}

func getKubeClient(t *testing.T) *kube.Client {
	config := &kube_client.Config{
		Host:     "http://0.0.0.0:8080",
		Insecure: false,
	}
	k, err := kube.New(config)
	require.NoError(t, err)
	return k
}

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}

func uniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}
