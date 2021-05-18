package testing

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
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

// Helper functions for tests below
func provStr(i interface{}) interface{} {
	cp := i.(*pfs.CommitProvenance)
	return fmt.Sprintf("%s@%s=%s", cp.Commit.Branch.Repo.Name, cp.Commit.Branch.Name, cp.Commit.ID)
}

func subvStr(i interface{}) interface{} {
	cr := i.(*pfs.CommitRange)
	return fmt.Sprintf("%s@%s=%s:%s@%s=%s", cr.Lower.Branch.Repo.Name, cr.Lower.Branch.Name, cr.Lower.ID, cr.Upper.Branch.Repo.Name, cr.Upper.Branch.Name, cr.Upper.ID)
}

func expectProv(commits ...*pfs.Commit) []interface{} {
	result := []interface{}{}
	for _, commit := range commits {
		result = append(result, provStr(commit.NewProvenance()))
	}
	return result
}

func expectSubv(commits ...*pfs.Commit) []interface{} {
	result := []interface{}{}
	for _, commit := range commits {
		result = append(result, subvStr(&pfs.CommitRange{Lower: commit, Upper: commit}))
	}
	return result
}

func TestTransactions(suite *testing.T) {
	suite.Parallel()

	suite.Run("TestEmptyTransaction", func(t *testing.T) {
		t.Parallel()
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		// Create repo, branch, start commit, finish commit
		_, err = txnClient.PfsAPIClient.CreateRepo(txnClient.Ctx(), &pfs.CreateRepoRequest{
			Repo: client.NewRepo("foo"),
		})
		require.NoError(t, err)

		_, err = txnClient.PfsAPIClient.CreateBranch(txnClient.Ctx(), &pfs.CreateBranchRequest{
			Branch: client.NewBranch("foo", "master")},
		)
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
		require.Equal(t, 4, len(info.Requests))
		require.Equal(t, 4, len(info.Responses))

		// Check each response value
		requireEmptyResponse(t, info.Responses[0])
		requireEmptyResponse(t, info.Responses[1])
		requireCommitResponse(t, info.Responses[2], commit)
		requireEmptyResponse(t, info.Responses[3])

		info, err = env.PachClient.FinishTransaction(txn)
		require.NoError(t, err)
		require.Equal(t, txn, info.Transaction)
		require.Equal(t, 4, len(info.Requests))
		require.Equal(t, 4, len(info.Responses))

		// Double-check each response value
		requireEmptyResponse(t, info.Responses[0])
		requireEmptyResponse(t, info.Responses[1])
		requireCommitResponse(t, info.Responses[2], commit)
		requireEmptyResponse(t, info.Responses[3])

		info, err = env.PachClient.InspectTransaction(txn)
		require.YesError(t, err)
		require.Nil(t, info)
	})

	// This is a regression test for a bug in PFS where creating a branch would
	// inspect the new commit outside of the transaction STM and fail to find it.
	suite.Run("TestCreateBranch", func(t *testing.T) {
		t.Parallel()
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		err = txnClient.CreateRepo("foo")
		require.NoError(t, err)

		commit1, err := txnClient.StartCommit("foo", "master")
		require.NoError(t, err)
		err = txnClient.FinishCommit("foo", "master", "")
		require.NoError(t, err)

		commit2, err := txnClient.StartCommit("foo", "master")
		require.NoError(t, err)
		err = txnClient.FinishCommit("foo", "master", "")
		require.NoError(t, err)

		require.NotEqual(t, commit1, commit2)

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		// Double-check each response value
		requireEmptyResponse(t, info.Responses[0])
		requireCommitResponse(t, info.Responses[1], commit1)
		requireEmptyResponse(t, info.Responses[2])
		requireCommitResponse(t, info.Responses[3], commit2)
		requireEmptyResponse(t, info.Responses[4])
	})

	// Test that a transactional change to multiple repos will only propagate a
	// single commit into a downstream repo. This mimics the pfs.TestProvenance test
	// using the following DAG:
	//  A ─▶ B ─▶ C ─▶ D
	//            ▲
	//  E ────────╯
	suite.Run("TestPropagateCommit", func(t *testing.T) {
		t.Parallel()
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

		require.NoError(t, env.PachClient.CreateRepo("A"))
		require.NoError(t, env.PachClient.CreateRepo("B"))
		require.NoError(t, env.PachClient.CreateRepo("C"))
		require.NoError(t, env.PachClient.CreateRepo("D"))
		require.NoError(t, env.PachClient.CreateRepo("E"))

		require.NoError(t, env.PachClient.CreateBranch("B", "master", "", "", []*pfs.Branch{client.NewBranch("A", "master")}))
		require.NoError(t, env.PachClient.CreateBranch("C", "master", "", "", []*pfs.Branch{client.NewBranch("B", "master"), client.NewBranch("E", "master")}))
		require.NoError(t, env.PachClient.CreateBranch("D", "master", "", "", []*pfs.Branch{client.NewBranch("C", "master")}))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		commitA, err := txnClient.StartCommit("A", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("A", "master", ""))
		commitE, err := txnClient.StartCommit("E", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("E", "master", ""))

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		require.Equal(t, 4, len(info.Responses))
		requireCommitResponse(t, info.Responses[0], commitA)
		requireEmptyResponse(t, info.Responses[1])
		requireCommitResponse(t, info.Responses[2], commitE)
		requireEmptyResponse(t, info.Responses[3])

		commitInfos, err := env.PachClient.ListCommitByRepo(client.NewRepo("A"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoA := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("B"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoB := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("C"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoC := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("D"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoD := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("E"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoE := commitInfos[0]

		require.Equal(t, commitA, commitInfoA.Commit)
		commitB := commitInfoB.Commit
		commitC := commitInfoC.Commit
		commitD := commitInfoD.Commit
		require.Equal(t, commitE, commitInfoE.Commit)

		require.ElementsEqualUnderFn(t, expectProv(), commitInfoA.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitB, commitC, commitD), commitInfoA.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(commitA), commitInfoB.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitC, commitD), commitInfoB.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(commitA, commitB, commitE), commitInfoC.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitD), commitInfoC.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(commitA, commitB, commitC, commitE), commitInfoD.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(), commitInfoD.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(), commitInfoE.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitC, commitD), commitInfoE.Subvenance, subvStr)
	})

	// This test is the same as PropagateCommit except more of the operations are
	// performed within the transaction.
	suite.Run("TestPropagateCommitRedux", func(t *testing.T) {
		t.Parallel()
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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

		commitA, err := txnClient.StartCommit("A", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("A", "master", ""))
		commitE, err := txnClient.StartCommit("E", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("E", "master", ""))

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		require.Equal(t, 12, len(info.Responses))
		requireCommitResponse(t, info.Responses[8], commitA)
		requireEmptyResponse(t, info.Responses[9])
		requireCommitResponse(t, info.Responses[10], commitE)
		requireEmptyResponse(t, info.Responses[11])

		commitInfos, err := env.PachClient.ListCommitByRepo(client.NewRepo("A"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoA := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("B"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoB := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("C"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoC := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("D"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoD := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommitByRepo(client.NewRepo("E"))
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoE := commitInfos[0]

		require.Equal(t, commitA, commitInfoA.Commit)
		commitB := commitInfoB.Commit
		commitC := commitInfoC.Commit
		commitD := commitInfoD.Commit
		require.Equal(t, commitE, commitInfoE.Commit)

		require.ElementsEqualUnderFn(t, expectProv(), commitInfoA.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitB, commitC, commitD), commitInfoA.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(commitA), commitInfoB.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitC, commitD), commitInfoB.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(commitA, commitB, commitE), commitInfoC.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitD), commitInfoC.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(commitA, commitB, commitC, commitE), commitInfoD.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(), commitInfoD.Subvenance, subvStr)

		require.ElementsEqualUnderFn(t, expectProv(), commitInfoE.Provenance, provStr)
		require.ElementsEqualUnderFn(t, expectSubv(commitC, commitD), commitInfoE.Subvenance, subvStr)
	})

	suite.Run("TestBatchTransaction", func(t *testing.T) {
		t.Parallel()
		env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))

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
			require.NoError(t, builder.CreateBranch("repoB", "branchA", "master", "", []*pfs.Branch{}))
			require.NoError(t, builder.CreateBranch("repoB", "branchB", "branchA", "", []*pfs.Branch{}))
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 4, len(info.Requests))
		require.Equal(t, 4, len(info.Responses))

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
	c := testutil.GetPachClient(t)
	require.NoError(t, c.DeleteAll())
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
	c.PutFile(commit, "foo", strings.NewReader("bar"))
	commitInfos, err := c.FlushCommitAll([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(commitInfos))

	var buf bytes.Buffer
	require.NoError(t, c.GetFile(commitInfos[0].Commit, "foo", &buf))
	require.Equal(t, "bar", buf.String())
}
