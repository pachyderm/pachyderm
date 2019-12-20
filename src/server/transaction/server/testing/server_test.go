package testing

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestEmptyTransaction(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
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

		return nil
	})
	require.NoError(t, err)
}

func TestInvalidatedTransaction(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
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

		return nil
	})
	require.NoError(t, err)
}

func TestFailedAppend(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
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

		return nil
	})
	require.NoError(t, err)
}

func requireEmptyResponse(t *testing.T, response *transaction.TransactionResponse) {
	require.Nil(t, response.Commit)
}

func requireCommitResponse(t *testing.T, response *transaction.TransactionResponse, commit *pfs.Commit) {
	require.Equal(t, commit, response.Commit)
}

func TestDependency(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
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
			Branch: "master",
			Parent: client.NewCommit("foo", ""),
		})
		require.NoError(t, err)

		_, err = txnClient.PfsAPIClient.FinishCommit(txnClient.Ctx(), &pfs.FinishCommitRequest{
			Commit: client.NewCommit("foo", "master"),
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

		return nil
	})
	require.NoError(t, err)
}

func TestDeleteAllTransactions(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
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

		return nil
	})
	require.NoError(t, err)
}

func TestMultiCommit(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		err = txnClient.CreateRepo("foo")
		require.NoError(t, err)

		commit1, err := txnClient.StartCommit("foo", "master")
		require.NoError(t, err)
		err = txnClient.FinishCommit("foo", "master")
		require.NoError(t, err)

		commit2, err := txnClient.StartCommit("foo", "master")
		require.NoError(t, err)
		err = txnClient.FinishCommit("foo", "master")
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

		return nil
	})
	require.NoError(t, err)
}

// Helper functions for tests below
func provStr(i interface{}) interface{} {
	cp := i.(*pfs.CommitProvenance)
	return fmt.Sprintf("%s@%s (%s)", cp.Commit.Repo.Name, cp.Commit.ID, cp.Branch.Name)
}

func subvStr(i interface{}) interface{} {
	cr := i.(*pfs.CommitRange)
	return fmt.Sprintf("%s@%s:%s@%s", cr.Lower.Repo.Name, cr.Lower.ID, cr.Upper.Repo.Name, cr.Upper.ID)
}

func expectProv(commits ...*pfs.Commit) []interface{} {
	result := []interface{}{}
	for _, commit := range commits {
		result = append(result, provStr(client.NewCommitProvenance(commit.Repo.Name, "master", commit.ID)))
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

// Test that a transactional change to multiple repos will only propagate a
// single commit into a downstream repo. This mimics the pfs.TestProvenance test
// using the following DAG:
//  A ─▶ B ─▶ C ─▶ D
//            ▲
//  E ────────╯
func TestPropagateCommit(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
		require.NoError(t, env.PachClient.CreateRepo("A"))
		require.NoError(t, env.PachClient.CreateRepo("B"))
		require.NoError(t, env.PachClient.CreateRepo("C"))
		require.NoError(t, env.PachClient.CreateRepo("D"))
		require.NoError(t, env.PachClient.CreateRepo("E"))

		require.NoError(t, env.PachClient.CreateBranch("B", "master", "", []*pfs.Branch{client.NewBranch("A", "master")}))
		require.NoError(t, env.PachClient.CreateBranch("C", "master", "", []*pfs.Branch{client.NewBranch("B", "master"), client.NewBranch("E", "master")}))
		require.NoError(t, env.PachClient.CreateBranch("D", "master", "", []*pfs.Branch{client.NewBranch("C", "master")}))

		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		commitA, err := txnClient.StartCommit("A", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("A", "master"))
		commitE, err := txnClient.StartCommit("E", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("E", "master"))

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		require.Equal(t, 4, len(info.Responses))
		requireCommitResponse(t, info.Responses[0], commitA)
		requireEmptyResponse(t, info.Responses[1])
		requireCommitResponse(t, info.Responses[2], commitE)
		requireEmptyResponse(t, info.Responses[3])

		commitInfos, err := env.PachClient.ListCommit("A", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoA := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("B", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoB := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("C", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoC := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("D", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoD := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("E", "", "", 0)
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

		return nil
	})
	require.NoError(t, err)
}

// This test is the same as PropagateCommit except more of the operations are
// performed within the transaction.
func TestPropagateCommitRedux(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
		txn, err := env.PachClient.StartTransaction()
		require.NoError(t, err)

		txnClient := env.PachClient.WithTransaction(txn)

		require.NoError(t, txnClient.CreateRepo("A"))
		require.NoError(t, txnClient.CreateRepo("B"))
		require.NoError(t, txnClient.CreateRepo("C"))
		require.NoError(t, txnClient.CreateRepo("D"))
		require.NoError(t, txnClient.CreateRepo("E"))

		require.NoError(t, txnClient.CreateBranch("B", "master", "", []*pfs.Branch{client.NewBranch("A", "master")}))
		require.NoError(t, txnClient.CreateBranch("C", "master", "", []*pfs.Branch{client.NewBranch("B", "master"), client.NewBranch("E", "master")}))
		require.NoError(t, txnClient.CreateBranch("D", "master", "", []*pfs.Branch{client.NewBranch("C", "master")}))

		commitA, err := txnClient.StartCommit("A", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("A", "master"))
		commitE, err := txnClient.StartCommit("E", "master")
		require.NoError(t, err)
		require.NoError(t, txnClient.FinishCommit("E", "master"))

		info, err := txnClient.FinishTransaction(txn)
		require.NoError(t, err)

		require.Equal(t, 12, len(info.Responses))
		requireCommitResponse(t, info.Responses[8], commitA)
		requireEmptyResponse(t, info.Responses[9])
		requireCommitResponse(t, info.Responses[10], commitE)
		requireEmptyResponse(t, info.Responses[11])

		commitInfos, err := env.PachClient.ListCommit("A", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoA := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("B", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoB := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("C", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoC := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("D", "", "", 0)
		require.NoError(t, err)
		require.Equal(t, 1, len(commitInfos))
		commitInfoD := commitInfos[0]

		commitInfos, err = env.PachClient.ListCommit("E", "", "", 0)
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

		return nil
	})
	require.NoError(t, err)
}

func TestBatchTransaction(t *testing.T) {
	t.Parallel()
	err := tu.WithRealEnv(func(env *tu.RealEnv) error {
		var branches []*pfs.BranchInfo
		var info *transaction.TransactionInfo
		var err error

		getBranchNames := func(branches []*pfs.BranchInfo) (result []string) {
			for _, branch := range branches {
				result = append(result, branch.Name)
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
			require.NoError(t, builder.CreateBranch("repoA", "master", "", []*pfs.Branch{}))
			require.NoError(t, builder.CreateBranch("repoA", "branchA", "master", []*pfs.Branch{}))
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 2, len(info.Requests))
		require.Equal(t, 2, len(info.Responses))

		branches, err = env.PachClient.ListBranch("repoA")
		require.NoError(t, err)

		require.ElementsEqual(t, []string{"master", "branchA"}, getBranchNames(branches))

		// Some dependent operations
		info, err = env.PachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
			require.NoError(t, builder.CreateRepo("repoB"))
			_, err := builder.StartCommit("repoB", "master")
			require.NoError(t, err)
			require.NoError(t, builder.CreateBranch("repoB", "branchA", "master", []*pfs.Branch{}))
			require.NoError(t, builder.CreateBranch("repoB", "branchB", "branchA", []*pfs.Branch{}))
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, 4, len(info.Requests))
		require.Equal(t, 4, len(info.Responses))

		branches, err = env.PachClient.ListBranch("repoB")
		require.NoError(t, err)

		require.ElementsEqual(t, []string{"master", "branchA", "branchB"}, getBranchNames(branches))

		// TODO: check that the result of the startcommit matches the master branch

		return nil
	})
	require.NoError(t, err)
}
