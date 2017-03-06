package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	pfspretty "github.com/pachyderm/pachyderm/src/server/pfs/pretty"
	ppspretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"
	pps_server "github.com/pachyderm/pachyderm/src/server/pps/server"

	"k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/restclient"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

func TestRestartOne(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// this test cannot be run in parallel because it restarts everything which breaks other tests.
	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestRestartOne_data")
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
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/",
		}},
		"",
		false,
	))
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	collectCommitInfos(t, commitIter)

	restartOne(t)

	// need a new client because the old one will have a defunct connection
	c = getUsablePachClient(t)

	_, err = c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	_, err = c.InspectRepo(dataRepo)
	require.NoError(t, err)
	_, err = c.InspectCommit(dataRepo, commit.ID)
	require.NoError(t, err)
}

func TestPrettyPrinting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPrettyPrinting_data")
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
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))
	// Do a commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	repoInfo, err := c.InspectRepo(dataRepo)
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedRepoInfo(repoInfo))
	for _, commitInfo := range commitInfos {
		require.NoError(t, pfspretty.PrintDetailedCommitInfo(commitInfo))
	}
	fileInfo, err := c.InspectFile(dataRepo, commit.ID, "file")
	require.NoError(t, err)
	require.NoError(t, pfspretty.PrintDetailedFileInfo(fileInfo))
	pipelineInfo, err := c.InspectPipeline(pipelineName)
	require.NoError(t, err)
	require.NoError(t, ppspretty.PrintDetailedPipelineInfo(pipelineInfo))
	jobInfos, err := c.ListJob("", nil)
	require.NoError(t, err)
	require.True(t, len(jobInfos) > 0)
	require.NoError(t, ppspretty.PrintDetailedJobInfo(jobInfos[0]))
}

func TestDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// this test cannot be run in parallel because it deletes everything
	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestDeleteAll_data")
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
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/",
		}},
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(collectCommitInfos(t, commitIter)))
	require.NoError(t, c.DeleteAll())
	repoInfos, err := c.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))
	pipelineInfos, err := c.ListPipeline()
	require.NoError(t, err)
	require.Equal(t, 0, len(pipelineInfos))
	jobInfos, err := c.ListJob("", nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(jobInfos))
}

func TestRecursiveCp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()

	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestRecursiveCp_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("TestRecursiveCp")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"sh"},
		[]string{
			fmt.Sprintf("cp -r /pfs/%s /pfs/out", dataRepo),
		},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(dataRepo),
			Glob: "/*",
		}},
		"",
		false,
	))
	// Do commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	for i := 0; i < 100; i++ {
		_, err = c.PutFile(
			dataRepo,
			commit.ID,
			fmt.Sprintf("file%d", i),
			strings.NewReader(strings.Repeat("foo\n", 10000)),
		)
		require.NoError(t, err)
	}
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(collectCommitInfos(t, commitIter)))
}

func TestPipelineUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	repo := uniqueString("data")
	require.NoError(t, c.CreateRepo(repo))
	pipelineName := uniqueString("pipeline")
	require.NoError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{
			{
				Repo: &pfs.Repo{Name: repo},
				Glob: "/",
			},
		},
		"",
		false,
	))
	err := c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{
			{
				Repo: &pfs.Repo{Name: repo},
				Glob: "/",
			},
		},
		"",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "pipeline .*? already exists", err.Error())
}

//func TestUpdatePipeline(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()
//c := getPachClient(t)
//// create repos
//dataRepo := uniqueString("TestUpdatePipeline_data")
//require.NoError(t, c.CreateRepo(dataRepo))
//// create 2 pipelines
//pipelineName := uniqueString("pipeline")
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file1 >>/pfs/out/file
//`, dataRepo)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{
//Repo: client.NewRepo(dataRepo),
//Glob: "/*",
//}},
//"",
//false,
//))
//pipeline2Name := uniqueString("pipeline2")
//require.NoError(t, c.CreatePipeline(
//pipeline2Name,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file >>/pfs/out/file
//`, pipelineName)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{
//Repo: client.NewRepo(pipelineName),
//Glob: "/*",
//}},
//"",
//false,
//))
//// Do first commit to repo
//var commit *pfs.Commit
//var err error
//for i := 0; i < 2; i++ {
//commit, err = c.StartCommit(dataRepo, "")
//require.NoError(t, err)
//if i == 0 {
//require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
//}
//_, err = c.PutFile(dataRepo, commit.ID, "file1", strings.NewReader("file1\n"))
//_, err = c.PutFile(dataRepo, commit.ID, "file2", strings.NewReader("file2\n"))
//_, err = c.PutFile(dataRepo, commit.ID, "file3", strings.NewReader("file3\n"))
//require.NoError(t, err)
//require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
//}
//commitInfos, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//require.Equal(t, 3, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "file1\nfile1\n", buffer.String())
//}

//// We archive the temporary commits created per job/pod
//// So the total we see here is 4, but 'real' commits is just 1
//outputRepoCommitInfos, err := c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
//require.NoError(t, err)
//require.Equal(t, 4, len(outputRepoCommitInfos))

//outputRepoCommitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))

//// Update the pipeline to look at file2
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file2 >>/pfs/out/file
//`, dataRepo)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{Repo: &pfs.Repo{Name: dataRepo}}},
//true,
//))
//pipelineInfo, err := c.InspectPipeline(pipelineName)
//require.NoError(t, err)
//require.NotNil(t, pipelineInfo.CreatedAt)
//commitInfos, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//require.Equal(t, 3, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "file2\nfile2\n", buffer.String())
//}
//outputRepoCommitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
//require.NoError(t, err)
//require.Equal(t, 8, len(outputRepoCommitInfos))
//// Expect real commits to still be 1
//outputRepoCommitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))

//// Update the pipeline to look at file3
//require.NoError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{fmt.Sprintf(`
//cat /pfs/%s/file3 >>/pfs/out/file
//`, dataRepo)},
//&pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*pps.PipelineInput{{Repo: &pfs.Repo{Name: dataRepo}}},
//true,
//))
//commitInfos, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//require.Equal(t, 3, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "file3\nfile3\n", buffer.String())
//}
//outputRepoCommitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
//require.NoError(t, err)
//require.Equal(t, 12, len(outputRepoCommitInfos))
//// Expect real commits to still be 1
//outputRepoCommitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))

//commitInfos, _ = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
//// Do an update that shouldn't cause archiving
//_, err = c.PpsAPIClient.CreatePipeline(
//context.Background(),
//&pps.CreatePipelineRequest{
//Pipeline: client.NewPipeline(pipelineName),
//Transform: &pps.Transform{
//Cmd: []string{"bash"},
//Stdin: []string{fmt.Sprintf(`
//cat /pfs/%s/file3 >>/pfs/out/file
//`, dataRepo)},
//},
//ParallelismSpec: &pps.ParallelismSpec{
//Strategy: pps.ParallelismSpec_CONSTANT,
//Constant: 2,
//},
//Inputs:    []*pps.PipelineInput{{Repo: &pfs.Repo{Name: dataRepo}}},
//Update:    true,
//NoArchive: true,
//})
//require.NoError(t, err)
//commitInfos, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
//require.NoError(t, err)
//require.Equal(t, 3, len(commitInfos))
//// only care about non-provenance commits
//commitInfos = commitInfos[1:]
//for _, commitInfo := range commitInfos {
//var buffer bytes.Buffer
//require.NoError(t, c.GetFile(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID, "file", 0, 0, "", false, nil, &buffer))
//require.Equal(t, "file3\nfile3\n", buffer.String())
//}
//commitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusAll, false)
//require.NoError(t, err)
//require.Equal(t, 12, len(commitInfos))
//// Expect real commits to still be 1
//outputRepoCommitInfos, err = c.ListCommitByRepo([]string{pipelineName}, nil, client.CommitTypeRead, client.CommitStatusNormal, false)
//require.NoError(t, err)
//require.Equal(t, 2, len(outputRepoCommitInfos))
//}

func TestStopPipeline(t *testing.T) {
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
			Repo: &pfs.Repo{Name: dataRepo},
			Glob: "/*",
		}},
		"",
		false,
	))
	require.NoError(t, c.StopPipeline(pipelineName))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	// wait for 10 seconds and check that no commit has been outputted
	time.Sleep(10 * time.Second)
	commits, err := c.ListCommit(pipelineName, "", "", 0)
	require.NoError(t, err)
	require.Equal(t, len(commits), 0)
	require.NoError(t, c.StartPipeline(pipelineName))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
}

func TestPipelineEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()

	// make a secret to reference
	k := getKubeClient(t)
	secretName := uniqueString("test-secret")
	_, err := k.Secrets(api.NamespaceDefault).Create(
		&api.Secret{
			ObjectMeta: api.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				"foo": []byte("foo\n"),
			},
		},
	)
	require.NoError(t, err)
	c := getPachClient(t)
	// create repos
	dataRepo := uniqueString("TestPipelineEnv_data")
	require.NoError(t, c.CreateRepo(dataRepo))
	// create pipeline
	pipelineName := uniqueString("pipeline")
	_, err = c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"sh"},
				Stdin: []string{
					"ls /var/secret",
					"cat /var/secret/foo > /pfs/out/foo",
					"echo $bar> /pfs/out/bar",
				},
				Env: map[string]string{"bar": "bar"},
				Secrets: []*pps.Secret{
					{
						Name:      secretName,
						MountPath: "/var/secret",
					},
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			Inputs: []*pps.PipelineInput{{
				Repo: &pfs.Repo{Name: dataRepo},
				Glob: "/*",
			}},
		})
	require.NoError(t, err)
	// Do first commit to repo
	commit, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit.ID, "master"))
	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	commitIter, err := c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "foo", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(pipelineName, commitInfos[0].Commit.ID, "bar", 0, 0, &buffer))
	require.Equal(t, "bar\n", buffer.String())
}

func TestPipelineWithFullObjects(t *testing.T) {
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
		[]*pps.PipelineInput{
			{
				Repo: client.NewRepo(dataRepo),
				Glob: "/*",
			},
		},
		"",
		false,
	))
	// Do first commit to repo
	commit1, err := c.StartCommit(dataRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dataRepo, commit1.ID, "master"))
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))
	commitInfoIter, err := c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, commit1.ID)}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	// Do second commit to repo
	commit2, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit2.ID, "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit2.ID))
	commitInfoIter, err = c.FlushCommit([]*pfs.Commit{client.NewCommit(dataRepo, "master")}, nil)
	require.NoError(t, err)
	commitInfos = collectCommitInfos(t, commitInfoIter)
	require.Equal(t, 1, len(commitInfos))
	buffer = bytes.Buffer{}
	require.NoError(t, c.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buffer))
	require.Equal(t, "foo\nbar\n", buffer.String())
}

// TestChainedPipelines tracks https://github.com/pachyderm/pachyderm/issues/797
func TestChainedPipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)
	aRepo := uniqueString("A")
	require.NoError(t, c.CreateRepo(aRepo))

	dRepo := uniqueString("D")
	require.NoError(t, c.CreateRepo(dRepo))

	aCommit, err := c.StartCommit(aRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(aRepo, aCommit.ID, "master"))
	_, err = c.PutFile(aRepo, "master", "file", strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(aRepo, "master"))

	dCommit, err := c.StartCommit(dRepo, "")
	require.NoError(t, err)
	require.NoError(t, c.SetBranch(dRepo, dCommit.ID, "master"))
	_, err = c.PutFile(dRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dRepo, "master"))

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
			fmt.Sprintf("cp /pfs/%s/file /pfs/out/dFile", dRepo)},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Repo: client.NewRepo(bPipeline),
			Glob: "/",
		}, {
			Repo: client.NewRepo(dRepo),
			Glob: "/",
		}},
		"",
		false,
	))
	resultIter, err := c.FlushCommit([]*pfs.Commit{aCommit, dCommit}, nil)
	require.NoError(t, err)
	results := collectCommitInfos(t, resultIter)
	require.Equal(t, 1, len(results))
}

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
			Repo: client.NewRepo(bPipeline),
			Glob: "/",
		}, {
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
			Repo: client.NewRepo(cPipeline),
			Glob: "/",
		}},
		"",
		false,
	))

	resultsIter, err := c.FlushCommit([]*pfs.Commit{aCommit, eCommit}, nil)
	require.NoError(t, err)
	results := collectCommitInfos(t, resultsIter)
	require.Equal(t, 2, len(results))

	eCommit2, err := c.StartCommit(eRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(eRepo, "master", "file", strings.NewReader("bar\n"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(eRepo, "master"))

	resultsIter, err = c.FlushCommit([]*pfs.Commit{eCommit2}, nil)
	require.NoError(t, err)
	results = collectCommitInfos(t, resultsIter)
	require.Equal(t, 2, len(results))

	// Get number of jobs triggered in pipeline D
	jobInfos, err := c.ListJob(dPipeline, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobInfos))
}

func collectCommitInfos(t *testing.T, commitInfoIter client.CommitInfoIterator) []*pfs.CommitInfo {
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

func restartOne(t *testing.T) {
	k := getKubeClient(t)
	podsInterface := k.Pods(api.NamespaceDefault)
	labelSelector, err := labels.Parse("app=pachd")
	require.NoError(t, err)
	podList, err := podsInterface.List(
		api.ListOptions{
			LabelSelector: labelSelector,
		})
	require.NoError(t, err)
	require.NoError(t, podsInterface.Delete(podList.Items[rand.Intn(len(podList.Items))].Name, api.NewDeleteOptions(0)))
	waitForReadiness(t)
}

const (
	retries = 10
)

// getUsablePachClient is like getPachClient except it blocks until it gets a
// connection that actually works
func getUsablePachClient(t *testing.T) *client.APIClient {
	for i := 0; i < retries; i++ {
		client := getPachClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel() //cleanup resources
		_, err := client.PfsAPIClient.ListRepo(ctx, &pfs.ListRepoRequest{})
		if err == nil {
			return client
		}
	}
	t.Fatalf("failed to connect after %d tries", retries)
	return nil
}

func waitForReadiness(t *testing.T) {
	k := getKubeClient(t)
	rc := pachdRc(t)
	for {
		// This code is taken from
		// k8s.io/kubernetes/pkg/client/unversioned.ControllerHasDesiredReplicas
		// It used to call that fun ction but an update to the k8s library
		// broke it due to a type error.  We should see if we can go back to
		// using that code but I(jdoliner) couldn't figure out how to fanagle
		// the types into compiling.
		newRc, err := k.ReplicationControllers(api.NamespaceDefault).Get(rc.Name)
		require.NoError(t, err)
		if newRc.Status.ObservedGeneration >= rc.Generation && newRc.Status.Replicas == newRc.Spec.Replicas {
			break
		}
		time.Sleep(time.Second * 5)
	}
	watch, err := k.Pods(api.NamespaceDefault).Watch(api.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "pachd"}),
	})
	defer watch.Stop()
	require.NoError(t, err)
	readyPods := make(map[string]bool)
	for event := range watch.ResultChan() {
		ready, err := kube.PodRunningAndReady(event)
		require.NoError(t, err)
		if ready {
			pod, ok := event.Object.(*api.Pod)
			if !ok {
				t.Fatal("event.Object should be an object")
			}
			readyPods[pod.Name] = true
			if len(readyPods) == int(rc.Spec.Replicas) {
				break
			}
		}
	}
}

func pachdRc(t *testing.T) *api.ReplicationController {
	k := getKubeClient(t)
	rc := k.ReplicationControllers(api.NamespaceDefault)
	result, err := rc.Get("pachd")
	require.NoError(t, err)
	return result
}
