//go:build k8s

package testing

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

func requireEmptyResponse(t *testing.T, response *transaction.TransactionResponse) {
	require.Nil(t, response.Commit)
}

func requireCommitResponse(t *testing.T, response *transaction.TransactionResponse, commit *pfs.Commit) {
	require.Equal(t, commit, response.Commit)
}

func TestTransactions(suite *testing.T) {
	suite.Parallel()

	suite.Run("TestEmptyTransaction", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		info, err := env.PachClient.InspectTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 0, len(info.Requests))
		require.Equal(t, 0, len(info.Responses))
		require.NotNil(t, info.Started)

		info, err = env.PachClient.FinishTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 0, len(info.Requests))
		require.Equal(t, 0, len(info.Responses))
		require.NotNil(t, info.Started)

		info, err = env.PachClient.InspectTransaction(txn)
		require.YesError(t, err)
		require.Nil(t, info)
	})

	suite.Run("TestInvalidatedTransaction", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)
		createRepo := &pfs.CreateRepoRequest{
			Repo: client.NewRepo("foo"),
		}

		// Tell the transaction to create a repo
		_, err = txnClient.PfsAPIClient.CreateRepo(txnClient.Ctx(), createRepo)
		require.NoError(t, err)

		// Create the same repo outside of the transaction, so it can't run
		_, err = env.PachClient.PfsAPIClient.CreateRepo(env.Context, createRepo)
		require.NoError(t, err)

		// Finishing the transaction should fail
		info, err := env.PachClient.FinishTransaction(txn)
		require.YesError(t, err)
		require.Nil(t, info)

		// Appending to the transaction should fail
		_, err = txnClient.PfsAPIClient.CreateRepo(txnClient.Ctx(), &pfs.CreateRepoRequest{Repo: client.NewRepo("bar")})
		require.YesError(t, err)
	})

	suite.Run("TestFailedAppend", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)
		createRepo := &pfs.CreateRepoRequest{
			Repo: client.NewRepo("foo"),
		}

		// Create a repo outside of a transaction
		_, err = env.PachClient.PfsAPIClient.CreateRepo(env.Context, createRepo)
		require.NoError(t, err)

		// Tell the transaction to create the same repo, which should fail
		_, err = txnClient.PfsAPIClient.CreateRepo(txnClient.Ctx(), createRepo)
		require.YesError(t, err)

		info, err := env.PachClient.InspectTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 0, len(info.Requests))
		require.Equal(t, 0, len(info.Responses))

		info, err = env.PachClient.FinishTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 0, len(info.Requests))
		require.Equal(t, 0, len(info.Responses))
	})

	suite.Run("TestDependency", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		// Create repo, start commit, finish commit
		_, err = txnClient.PfsAPIClient.CreateRepo(txnClient.Ctx(), &pfs.CreateRepoRequest{
			Repo: client.NewRepo("foo"),
		})
		require.NoError(t, err)

		commit, err := txnClient.PfsAPIClient.StartCommit(txnClient.Ctx(), &pfs.StartCommitRequest{
			Branch: client.NewBranch("foo", "master"),
		})
		require.NoError(t, err)

		_, err = txnClient.PfsAPIClient.FinishCommit(txnClient.Ctx(), &pfs.FinishCommitRequest{
			Commit: client.NewCommit("foo", "master", ""),
		})
		require.NoError(t, err)

		info, err := env.PachClient.InspectTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 3, len(info.Requests))
		require.Equal(t, 3, len(info.Responses))

		// Check each response value
		requireEmptyResponse(t, info.Responses[0])
		requireCommitResponse(t, info.Responses[1], commit)
		requireEmptyResponse(t, info.Responses[2])

		info, err = env.PachClient.FinishTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 3, len(info.Requests))
		require.Equal(t, 3, len(info.Responses))

		// Double-check each response value
		requireEmptyResponse(t, info.Responses[0])
		requireCommitResponse(t, info.Responses[1], commit)
		requireEmptyResponse(t, info.Responses[2])

		info, err = env.PachClient.InspectTransaction(txn)
		require.YesError(t, err)
		require.Nil(t, info)
	})

	// This is a regression test for a bug in PFS where creating a branch would
	// inspect the new commit outside of the transaction STM and fail to find it.
	suite.Run("TestCreateBranch", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		repo := "foo"
		branchA := "master"
		branchB := "bar"

		require.NoError(t, env.PachClient.CreateRepo(repo))
		require.NoError(t, env.PachClient.CreateBranch(repo, branchA, "", "", nil))
		require.NoError(t, env.PachClient.CreateBranch(repo, branchB, "", "", nil))

		txnClient := env.PachClient.WithTransaction(txn)
		commit, err := txnClient.StartCommit(repo, branchB)
		require.NoError(t, err)
		err = txnClient.FinishCommit(repo, branchB, "")
		require.NoError(t, err)
		require.NoError(t, txnClient.CreateBranch(repo, branchA, branchB, "", nil))

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		// Double-check each response value
		requireCommitResponse(t, info.Responses[0], commit)
		requireEmptyResponse(t, info.Responses[1])

		commitInfo, err := env.PachClient.InspectCommit(repo, branchA, "")
		require.NoError(t, err)
		require.Equal(t, commitInfo.Commit.ID, commit.ID)

		commitInfo, err = env.PachClient.InspectCommit(repo, branchB, "")
		require.NoError(t, err)
		require.Equal(t, commitInfo.Commit.ID, commit.ID)
	})

	suite.Run("TestDeleteAllTransactions", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		_, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		_, err = env.PachClient.StartTransaction()
		require.NoError(t, err)

		txns, err := env.PachClient.ListTransaction()
		require.NoError(t, err)
		require.Equal(t, 2, len(txns))

		_, err = env.PachClient.TransactionAPIClient.DeleteAll(env.Context, &transaction.DeleteAllRequest{})
		require.NoError(t, err)

		txns, err = env.PachClient.ListTransaction()
		require.NoError(t, err)
		require.Equal(t, 0, len(txns))
	})

	suite.Run("TestMultiCommit", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		err = txnClient.CreateRepo("foo")
		require.NoError(t, err)

		_, err = txnClient.StartCommit("foo", "master")
		require.NoError(t, err)
		err = txnClient.FinishCommit("foo", "master", "")
		require.NoError(t, err)

		_, err = txnClient.StartCommit("foo", "master")
		require.YesError(t, err)
		require.Matches(t, "already has a commit in this transaction", err.Error())
	})

	// Test that a transactional change to multiple repos will only propagate a
	// single commit into a downstream repo. This mimics the pfs.TestProvenance test
	// using the following DAG:
	//  A ─▶ B ─▶ C ─▶ D
	//            ▲
	//  E ────────╯
	suite.Run("TestPropagateCommit", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo("A"))
		require.NoError(t, env.PachClient.CreateRepo("B"))
		require.NoError(t, env.PachClient.CreateRepo("C"))
		require.NoError(t, env.PachClient.CreateRepo("D"))
		require.NoError(t, env.PachClient.CreateRepo("E"))

		require.NoError(t, env.PachClient.CreateBranch("B", "master", "", "", []*pfs.Branch{client.NewBranch("A", "master")}))
		require.NoError(t, env.PachClient.CreateBranch("C", "master", "", "", []*pfs.Branch{client.NewBranch("B", "master"), client.NewBranch("E", "master")}))
		require.NoError(t, env.PachClient.CreateBranch("D", "master", "", "", []*pfs.Branch{client.NewBranch("C", "master")}))

		commitInfos, err := env.PachClient.ListCommitByRepo(client.NewRepo("A"))
		require.NoError(t, err)
		aCommits := len(commitInfos)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("B"))
		require.NoError(t, err)
		bCommits := len(commitInfos)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("C"))
		require.NoError(t, err)
		cCommits := len(commitInfos)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("D"))
		require.NoError(t, err)
		dCommits := len(commitInfos)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("E"))
		require.NoError(t, err)
		eCommits := len(commitInfos)

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		commitA, err := txnClient.StartCommit("A", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("A", "master", ""))
		require.Equal(t, txn.ID, commitA.ID)
		commitE, err := txnClient.StartCommit("E", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("E", "master", ""))
		require.Equal(t, txn.ID, commitE.ID)

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		require.Equal(t, 4, len(info.Responses))
		requireCommitResponse(t, info.Responses[0], commitA)
		requireEmptyResponse(t, info.Responses[1])
		requireCommitResponse(t, info.Responses[2], commitE)
		requireEmptyResponse(t, info.Responses[3])

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("A"))
		require.NoError(t, err)
		require.Equal(t, aCommits+1, len(commitInfos))
		commitInfoA := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("B"))
		require.NoError(t, err)
		require.Equal(t, bCommits+1, len(commitInfos))

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("C"))
		require.NoError(t, err)
		require.Equal(t, cCommits+1, len(commitInfos))

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("D"))
		require.NoError(t, err)
		require.Equal(t, dCommits+1, len(commitInfos))

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("E"))
		require.NoError(t, err)
		require.Equal(t, eCommits+1, len(commitInfos))
		commitInfoE := commitInfos[0]

		require.Equal(t, commitA, commitInfoA.Commit)
		require.Equal(t, commitE, commitInfoE.Commit)
	})

	// This test is the same as PropagateCommit except more of the operations are
	// performed within the transaction.
	suite.Run("TestPropagateCommitRedux", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		require.NoError(t, txnClient.CreateRepo("A"))
		require.NoError(t, txnClient.CreateRepo("B"))
		require.NoError(t, txnClient.CreateRepo("C"))
		require.NoError(t, txnClient.CreateRepo("D"))
		require.NoError(t, txnClient.CreateRepo("E"))

		require.NoError(t, txnClient.CreateBranch("B", "master", "", "", []*pfs.Branch{client.NewBranch("A", "master")}))
		require.NoError(t, txnClient.CreateBranch("C", "master", "", "", []*pfs.Branch{client.NewBranch("B", "master"), client.NewBranch("E", "master")}))
		require.NoError(t, txnClient.CreateBranch("D", "master", "", "", []*pfs.Branch{client.NewBranch("C", "master")}))

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, 8, len(info.Responses))

		commitInfos, err := env.PachClient.ListCommitByRepo(client.NewRepo("A"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, txn.ID, commitInfos[0].Commit.ID)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("B"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, txn.ID, commitInfos[0].Commit.ID)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("C"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, txn.ID, commitInfos[0].Commit.ID)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("D"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, txn.ID, commitInfos[0].Commit.ID)

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("E"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		require.Equal(t, txn.ID, commitInfos[0].Commit.ID)
	})

	suite.Run("TestBatchTransaction", func(t *testing.T) {
		t.Parallel()
		env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))

		var branchInfos []*pfs.BranchInfo
		var info *transaction.TransactionInfo
		var err error

		getBranchNames := func(branchInfos []*pfs.BranchInfo) (result []string) {
			for _, branchInfo := range branchInfos {
				result = append(result, branchInfo.Branch.Name)
			}
			return result
		}

		// Empty batch
		info, err = env.PachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 0, len(info.Requests))
		require.Equal(t, 0, len(info.Responses))

		// One operation
		info, err = env.PachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
			require.NoError(t, builder.CreateRepo("repoA"))
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 1, len(info.Requests))
		require.Equal(t, 1, len(info.Responses))

		// Two independent operations
		info, err = env.PachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
			require.NoError(t, builder.CreateBranch("repoA", "master", "", "", []*pfs.Branch{}))
			require.NoError(t, builder.CreateBranch("repoA", "branchA", "master", "", []*pfs.Branch{}))
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 2, len(info.Requests))
		require.Equal(t, 2, len(info.Responses))

		branchInfos, err = env.PachClient.ListBranch("repoA")
		require.NoError(t, err)

		require.ElementsEqual(t, []string{"master", "branchA"}, getBranchNames(branchInfos))

		// Some dependent operations
		info, err = env.PachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
			require.NoError(t, builder.CreateRepo("repoB"))
			_, err := builder.StartCommit("repoB", "master")
			require.NoError(t, err)
			err = builder.FinishCommit("repoB", "master", "")
			require.NoError(t, err)
			require.NoError(t, builder.CreateBranch("repoB", "branchA", "master", "", []*pfs.Branch{}))
			require.NoError(t, builder.CreateBranch("repoB", "branchB", "branchA", "", []*pfs.Branch{}))
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 5, len(info.Requests))
		require.Equal(t, 5, len(info.Responses))

		branchInfos, err = env.PachClient.ListBranch("repoB")
		require.NoError(t, err)

		require.ElementsEqual(t, []string{"master", "branchA", "branchB"}, getBranchNames(branchInfos))

		for _, branchInfo := range branchInfos {
			if branchInfo.Branch.Name == "master" {
				require.Equal(t, branchInfo.Head, info.Responses[1].Commit)
			}
		}
	})
}

func TestCreatePipelineTransaction(t *testing.T) {
	c, _ := minikubetestenv.AcquireCluster(t)
	repo := testutil.UniqueString("in")
	pipeline := testutil.UniqueString("pipeline")
	_, err := c.ExecuteInTransaction(func(txnClient *client.APIClient) error {
		require.NoError(t, txnClient.CreateRepo(repo))
		require.NoError(t, txnClient.CreatePipeline(
			pipeline,
			"",
			[]string{"bash"},
			[]string{fmt.Sprintf("cp /pfs/%s/* /pfs/out", repo)},
			&pps.ParallelismSpec{Constant: 1},
			client.NewPFSInput(repo, "/"),
			"master",
			false,
		))
		return nil
	})
	require.NoError(t, err)

	commit := client.NewCommit(repo, "master", "")
	require.NoError(t, c.PutFile(commit, "foo", strings.NewReader("bar")))

	commitInfo, err := c.WaitCommit(pipeline, "master", "")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfo.Commit, "foo", &buf))
	require.Equal(t, "bar", buf.String())
}
