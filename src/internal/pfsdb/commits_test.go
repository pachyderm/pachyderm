package pfsdb_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestCreateCommitWithParent(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.NoError(t, err, "should be able to create parent commit")
			commitID, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			getInfo, err := pfsdb.GetCommit(ctx, tx, commitID)
			require.NoError(t, err, "should be able to get commit")
			require.Equal(t, getInfo.ParentCommit.Id, parentInfo.Commit.Id)
		})
	})
}

func TestCreateCommitWithMissingParent(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withFailedTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			_, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.YesError(t, err, "create commit should fail before creating parent")
			require.True(t, errors.Is(pfsdb.ErrParentCommitNotFound{ChildRowID: 1}, errors.Cause(err)))
		})
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			_, err := pfsdb.CreateCommit(ctx, tx, commitInfo, pfsdb.AncestryOpt{SkipParent: true})
			require.NoError(t, err, "should be able to create commit before creating parent")
		})
	})
}

func TestCreateCommitWithMissingChild(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withFailedTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo.ChildCommits = append(parentInfo.ChildCommits, commitInfo.Commit)
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.YesError(t, err, "create commit should fail before creating children")
			require.True(t, errors.Is(pfsdb.ErrChildCommitNotFound{ParentRowID: 1}, errors.Cause(err)))
		})
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo.ChildCommits = append(parentInfo.ChildCommits, commitInfo.Commit)
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo, pfsdb.AncestryOpt{SkipChildren: true})
			require.NoError(t, err, "should be able to create commit before creating children")
		})
	})
}

func TestCreateCommitWithRelatives(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			childInfo := testCommit(ctx, t, tx, testRepoName)
			childInfo2 := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			commitInfo.ChildCommits = append(commitInfo.ChildCommits, childInfo.Commit, childInfo2.Commit)
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.NoError(t, err, "should be able to create parent commit")
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo)
			require.NoError(t, err, "should be able to create child commit")
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo2)
			require.NoError(t, err, "should be able to create child commit 2")
			commitID, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			getInfo, err := pfsdb.GetCommit(ctx, tx, commitID)
			require.NoError(t, err, "should be able to get commit")
			require.Equal(t, getInfo.ParentCommit.Id, parentInfo.Commit.Id)
			require.Equal(t, getInfo.ChildCommits[0].Id, childInfo.Commit.Id)
			require.Equal(t, getInfo.ChildCommits[1].Id, childInfo2.Commit.Id)
		})
	})
}

func TestCreateCommit(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			createBranch(ctx, t, tx, commitInfo.Commit)
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, 1, commitInfo)) // add branch fields once they exist
			getInfo, err := pfsdb.GetCommit(ctx, tx, 1)
			require.NoError(t, err)
			commitsMatch(t, commitInfo, getInfo)
			commitInfo = testCommit(ctx, t, tx, testRepoName)
			commitInfo.Commit.Repo = nil
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.YesError(t, err, "should not be able to create commit when repo is nil")
			require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Repo"}, err))
			commitInfo = testCommit(ctx, t, tx, testRepoName)
			commitInfo.Origin = nil
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.YesError(t, err, "should not be able to create commit when origin is nil")
			require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Origin"}, err))
			commitInfo = testCommit(ctx, t, tx, testRepoName)
			commitInfo.Commit = nil
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.YesError(t, err, "should not be able to create commit when commit is nil")
			require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Commit"}, err))
		})
		withFailedTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create a commit")
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.YesError(t, err, "should not be able to create commit again with same commit ID")
			require.True(t, errors.Is(pfsdb.ErrCommitAlreadyExists{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, err))
		})
	})
}

func TestGetCommit(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			commitID, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			createBranch(ctx, t, tx, commitInfo.Commit)
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, commitID, commitInfo)) // add branch fields once they exist
			getInfo, err := pfsdb.GetCommit(ctx, tx, commitID)
			require.NoError(t, err, "should be able to get commit with id=1")
			commitsMatch(t, commitInfo, getInfo)
			_, err = pfsdb.GetCommit(ctx, tx, 0)
			require.YesError(t, err, "should not be able to get commit with id=0")
			_, err = pfsdb.GetCommit(ctx, tx, 3)
			require.YesError(t, err, "should not be able to get non-existent commit")
			require.True(t, errors.Is(pfsdb.ErrCommitNotFound{RowID: 3}, err))
			getInfo, err = pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to get a commit by key")
			commitsMatch(t, commitInfo, getInfo)
			commitInfo.Commit = nil
			_, err = pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.YesError(t, err, "should not be able to get commit when commit is nil.")
			require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Commit"}, errors.Cause(err)))
		})
	})
}

func TestDeleteCommitWithNoRelatives(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			err = pfsdb.DeleteCommit(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to delete commit")
			_, err = pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.YesError(t, err, "commit should no longer exist.")
			require.True(t, errors.Is(pfsdb.ErrCommitNotFound{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, errors.Cause(err)))
			err = pfsdb.DeleteCommit(ctx, tx, commitInfo.Commit)
			require.YesError(t, err, "should not be able to double delete commit")
			require.True(t, errors.Is(pfsdb.ErrCommitNotFound{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, errors.Cause(err)))
			commitInfo.Commit = nil
			err = pfsdb.DeleteCommit(ctx, tx, commitInfo.Commit)
			require.YesError(t, err, "should not be able to delete commitInfo when commit is missing")
			require.True(t, errors.Is(pfsdb.ErrCommitMissingInfo{Field: "Commit"}, errors.Cause(err)))
		})
	})
}

func TestDeleteCommitWithParent(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			// setup parent and commit
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.NoError(t, err, "should be able to create parent")
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			// validate parent and commit relationship
			parentID, err := pfsdb.GetCommitID(ctx, tx, parentInfo.Commit)
			require.NoError(t, err, "should be able to get parent commit row id")
			children, err := pfsdb.GetCommitChildren(ctx, tx, parentID)
			require.NoError(t, err, "should be able to get children of parent")
			require.Equal(t, len(children), 1, "there should only be 1 child")
			require.Equal(t, children[0].Id, commitInfo.Commit.Id, "commit should be parent's child")
			// delete commit
			err = pfsdb.DeleteCommit(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to delete commit")
			_, err = pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.YesError(t, err, "commit should not exist")
			require.True(t, errors.Is(pfsdb.ErrCommitNotFound{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, errors.Cause(err)))
			// confirm parent has no children.
			_, err = pfsdb.GetCommitChildren(ctx, tx, parentID)
			require.YesError(t, err, "should not be able to get any children.")
			require.Equal(t, pfsdb.ErrChildCommitNotFound{ParentRowID: parentID}, errors.Cause(err))
		})
	})
}

func TestDeleteCommitWithChildren(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			// setup parent and commit
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			childInfo := testCommit(ctx, t, tx, testRepoName)
			childInfo2 := testCommit(ctx, t, tx, testRepoName)
			childInfo.ParentCommit = commitInfo.Commit
			childInfo2.ParentCommit = commitInfo.Commit
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo)
			require.NoError(t, err, "should be able to create child")
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo2)
			require.NoError(t, err, "should be able to create child 2")
			// validate commit and children relationship
			childID, err := pfsdb.GetCommitID(ctx, tx, childInfo.Commit)
			require.NoError(t, err, "should be able to get child commit row id")
			childID2, err := pfsdb.GetCommitID(ctx, tx, childInfo2.Commit)
			require.NoError(t, err, "should be able to get child commit 2 row id")
			parent1, err := pfsdb.GetCommitParent(ctx, tx, childID)
			require.NoError(t, err, "should be able to get parent of child 1")
			parent2, err := pfsdb.GetCommitParent(ctx, tx, childID2)
			require.NoError(t, err, "should be able to get parent of child 2")
			require.Equal(t, parent1.Id, parent2.Id, "child 1 and child 2 should have same parent")
			require.Equal(t, parent1.Id, commitInfo.Commit.Id, "parent id should be commit id")
			// delete commit
			err = pfsdb.DeleteCommit(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to delete commit")
			_, err = pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.YesError(t, err, "commit should not exist")
			require.True(t, errors.Is(pfsdb.ErrCommitNotFound{CommitID: pfsdb.CommitKey(commitInfo.Commit)}, errors.Cause(err)))
			// confirm children has no parent.
			_, err = pfsdb.GetCommitParent(ctx, tx, childID)
			require.YesError(t, err, "parent of child 1 should not exist")
			require.True(t, errors.Is(pfsdb.ErrParentCommitNotFound{ChildRowID: childID}, errors.Cause(err)))
			_, err = pfsdb.GetCommitParent(ctx, tx, childID2)
			require.YesError(t, err, "parent of child 2 should not exist")
			require.True(t, errors.Is(pfsdb.ErrParentCommitNotFound{ChildRowID: childID2}, errors.Cause(err)))
		})
	})
}

func TestDeleteCommitWithRelatives(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			childInfo := testCommit(ctx, t, tx, testRepoName)
			childInfo2 := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			commitInfo.ChildCommits = append(commitInfo.ChildCommits, childInfo.Commit, childInfo2.Commit)
			childInfo.Finishing = timestamppb.New(time.Now())
			childInfo2.Finishing = timestamppb.New(time.Now())
			childInfo.Finished = timestamppb.New(time.Now())
			childInfo2.Finished = timestamppb.New(time.Now())
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.NoError(t, err, "should be able to create parent commit")
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo)
			require.NoError(t, err, "should be able to create child commit")
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo2)
			require.NoError(t, err, "should be able to create child commit 2")
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			// validate parent and commit relationship
			parentID, err := pfsdb.GetCommitID(ctx, tx, parentInfo.Commit)
			require.NoError(t, err, "should be able to get parent commit row id")
			children, err := pfsdb.GetCommitChildren(ctx, tx, parentID)
			require.NoError(t, err, "should be able to get children of parent")
			require.Equal(t, len(children), 1, "there should only be 1 child")
			require.Equal(t, children[0].Id, commitInfo.Commit.Id, "commit should be a child of parent")
			// validate commit and children relationship
			childID, err := pfsdb.GetCommitID(ctx, tx, childInfo.Commit)
			require.NoError(t, err, "should be able to get child commit row id")
			childID2, err := pfsdb.GetCommitID(ctx, tx, childInfo2.Commit)
			require.NoError(t, err, "should be able to get child 2 commit row id")
			parent1, err := pfsdb.GetCommitParent(ctx, tx, childID)
			require.NoError(t, err, "should be able to get parent of commit 1")
			parent2, err := pfsdb.GetCommitParent(ctx, tx, childID2)
			require.NoError(t, err, "should be able to get parent of commit 2")
			require.Equal(t, parent1.Id, parent2.Id, "child commit 1 and child commit 2 should have same parent")
			require.Equal(t, parent1.Id, commitInfo.Commit.Id, "commit should be parent of child commits")
			// delete commit
			err = pfsdb.DeleteCommit(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to delete commit")
			// confirm commit's children are now parent's children.
			children, err = pfsdb.GetCommitChildren(ctx, tx, parentID)
			require.NoError(t, err, "should able to get children.")
			require.Equal(t, len(children), 2, "parent should have 2 children")
			require.Equal(t, children[0].Id, childInfo.Commit.Id)
			require.Equal(t, children[1].Id, childInfo2.Commit.Id)
			// confirm children's parent is now parent.
			parent1, err = pfsdb.GetCommitParent(ctx, tx, childID)
			require.NoError(t, err, "should be able to get parent of child 1")
			parent2, err = pfsdb.GetCommitParent(ctx, tx, childID2)
			require.NoError(t, err, "should be able to get parent of child 2")
			require.Equal(t, parent1.Id, parent2.Id, "child commit 1 and child commit 2 should have same parent")
			require.Equal(t, parent1.Id, parentInfo.Commit.Id, "parent should be parent of child commits")
		})
	})
}

func TestUpdateCommitWithParent(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.NoError(t, err, "should be able to create parent commit")
			createBranch(ctx, t, tx, parentInfo.Commit)
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			id, err := pfsdb.GetCommitID(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to get commit id")
			parentInfo2 := testCommit(ctx, t, tx, testRepoName)
			_, err = pfsdb.CreateCommit(ctx, tx, parentInfo2)
			require.NoError(t, err, "should be able to create parent commit 2")
			commitInfo.Started = timestamppb.New(time.Now())
			commitInfo.ParentCommit = parentInfo2.Commit
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, id, commitInfo), "should be able to update commit")
			getInfo, err := pfsdb.GetCommit(ctx, tx, id)
			require.NoError(t, err, "should be able to get commit")
			commitsMatch(t, getInfo, commitInfo)
		})
	})
}

func TestUpdateCommitRemoveParent(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			parentInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ParentCommit = parentInfo.Commit
			_, err := pfsdb.CreateCommit(ctx, tx, parentInfo)
			require.NoError(t, err, "should be able to create parent")
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			commitInfo.ParentCommit = nil
			id, err := pfsdb.GetCommitID(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to get commit id")
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, id, commitInfo), "should be able to update commit")
			_, err = pfsdb.GetCommitParent(ctx, tx, id)
			require.YesError(t, err, "parent should not exist")
			require.True(t, errors.Is(pfsdb.ErrParentCommitNotFound{ChildRowID: id}, errors.Cause(err)))
		})
	})
}

func TestUpdateCommitWithChildren(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			childInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.CreateCommit(ctx, tx, childInfo)
			require.NoError(t, err, "should be able to create child commit")
			createBranch(ctx, t, tx, childInfo.Commit)
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ChildCommits = append(commitInfo.ChildCommits, childInfo.Commit)
			id, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			childInfo2 := testCommit(ctx, t, tx, testRepoName)
			_, err = pfsdb.CreateCommit(ctx, tx, childInfo2)
			require.NoError(t, err, "should be able to create child commit 2")
			commitInfo.Started = timestamppb.New(time.Now())
			commitInfo.ChildCommits = append(commitInfo.ChildCommits, childInfo2.Commit)
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, id, commitInfo), "should be able to update commit")
			getInfo, err := pfsdb.GetCommit(ctx, tx, id)
			require.NoError(t, err, "should be able to get commit")
			commitsMatch(t, getInfo, commitInfo)
		})
	})
}

func TestUpdateCommitRemoveChild(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			childInfo := testCommit(ctx, t, tx, testRepoName)
			commitInfo.ChildCommits = append(commitInfo.ChildCommits, childInfo.Commit)
			_, err := pfsdb.CreateCommit(ctx, tx, childInfo)
			require.NoError(t, err, "should be able to create child commit")
			_, err = pfsdb.CreateCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit")
			id, err := pfsdb.GetCommitID(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to get commit id")
			commitInfo.ChildCommits = make([]*pfs.Commit, 0)
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, id, commitInfo), "should be able to update commit")
			_, err = pfsdb.GetCommitChildren(ctx, tx, id)
			require.YesError(t, err, "children should not exist")
			require.True(t, errors.Is(pfsdb.ErrChildCommitNotFound{ParentRowID: id}, errors.Cause(err)))
		})
	})
}

func TestUpsertCommit(t *testing.T) {
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			commitInfo := testCommit(ctx, t, tx, testRepoName)
			_, err := pfsdb.UpsertCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to create commit via upsert")
			createBranch(ctx, t, tx, commitInfo.Commit)
			require.NoError(t, pfsdb.UpdateCommit(ctx, tx, 1, commitInfo)) // do an update to add the branch fields.
			getInfo, err := pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to get a commit by key")
			commitsMatch(t, commitInfo, getInfo)
			commitInfo.Started = timestamppb.New(time.Now())
			commitInfo.Description = "new desc"
			_, err = pfsdb.UpsertCommit(ctx, tx, commitInfo)
			require.NoError(t, err, "should be able to update commit via upsert")
			getInfo, err = pfsdb.GetCommitByCommitKey(ctx, tx, commitInfo.Commit)
			require.NoError(t, err, "should be able to get a commit by key")
			commitsMatch(t, commitInfo, getInfo)
		})
	})
}

func checkOutput(ctx context.Context, t *testing.T, iter stream.Iterator[pfsdb.CommitPair], expectedInfos []*pfs.CommitInfo) {
	i := 0
	require.NoError(t, stream.ForEach[pfsdb.CommitPair](ctx, iter, func(commitPair pfsdb.CommitPair) error {
		commitsMatch(t, expectedInfos[i], commitPair.CommitInfo)
		i++
		return nil
	}))
	require.Equal(t, len(expectedInfos), i)
}

func TestListCommit(t *testing.T) {
	size := 3
	expectedInfos := make([]*pfs.CommitInfo, size)
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			var prevCommit *pfs.CommitInfo
			for i := 0; i < size; i++ {
				commitInfo := testCommit(ctx, t, tx, testRepoName)
				if prevCommit != nil {
					commitInfo.ParentCommit = prevCommit.Commit
				}
				expectedInfos[i] = commitInfo
				commitID, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
				require.NoError(t, err, "should be able to create commit")
				createBranch(ctx, t, tx, commitInfo.Commit)
				if i == 0 { // the first commit will be missing branch information, so we need to add it.
					require.NoError(t, pfsdb.UpdateCommit(ctx, tx, 1, commitInfo))
				}
				if prevCommit != nil {
					require.NoError(t, pfsdb.CreateCommitParent(ctx, tx, prevCommit.Commit, commitID))
					expectedInfos[i-1].ChildCommits = append(expectedInfos[i-1].ChildCommits, commitInfo.Commit)
				}
				prevCommit = commitInfo
			}
		})
		iter, err := pfsdb.ListCommit(ctx, db, nil, false)
		require.NoError(t, err, "should be able to list repos")
		checkOutput(ctx, t, iter, expectedInfos)
	})
}

func TestListCommitRev(t *testing.T) {
	size := 210
	expectedInfos := make([]*pfs.CommitInfo, size)
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			var prevCommit *pfs.CommitInfo
			for i := size - 1; i >= 0; i-- {
				commitInfo := testCommit(ctx, t, tx, testRepoName)
				if prevCommit != nil {
					commitInfo.ParentCommit = prevCommit.Commit
				}
				expectedInfos[i] = commitInfo
				commitID, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
				require.NoError(t, err, "should be able to create commit")
				createBranch(ctx, t, tx, commitInfo.Commit)
				if i == size-1 { // the first commit will be missing branch information, so we need to add it.
					require.NoError(t, pfsdb.UpdateCommit(ctx, tx, 1, commitInfo))
				}
				if prevCommit != nil {
					require.NoError(t, pfsdb.CreateCommitParent(ctx, tx, prevCommit.Commit, commitID))
					expectedInfos[i+1].ChildCommits = append(expectedInfos[i+1].ChildCommits, commitInfo.Commit)
				}
				prevCommit = commitInfo
			}
		})
		iter, err := pfsdb.ListCommit(ctx, db, nil, true)
		require.NoError(t, err, "should be able to list repos")
		checkOutput(ctx, t, iter, expectedInfos)
	})
}

func TestListCommitsFilter(t *testing.T) {
	repos := []string{"a", "b", "c"}
	size := 330
	expectedInfos := make([]*pfs.CommitInfo, 0)
	commitSetIds := make([]string, 0)
	commits := make([]*pfs.CommitInfo, 0)
	filter := pfsdb.CommitListFilter{
		pfsdb.CommitRepos:    []string{"b"},
		pfsdb.CommitOrigins:  []string{pfs.OriginKind_ORIGIN_KIND_UNKNOWN.String(), pfs.OriginKind_USER.String()},
		pfsdb.CommitBranches: []string{"master"},
		pfsdb.CommitProjects: []string{"default"},
	}
	testCommitDataModelAPI(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			for i := 0; i < size; i++ {
				commitInfo := testCommit(ctx, t, tx, repos[i%len(repos)])
				if commitInfo.Commit.Repo.Name == "b" && i%10 == 0 {
					expectedInfos = append(expectedInfos, commitInfo)
					commitSetIds = append(commitSetIds, commitInfo.Commit.Id)
				}
				commits = append(commits, commitInfo)
			}
			filter[pfsdb.CommitSetIDs] = commitSetIds
			for _, commitInfo := range commits {
				_, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
				require.NoError(t, err, "should be able to create commit")
				createBranch(ctx, t, tx, commitInfo.Commit)
			}
		})
		iter, err := pfsdb.ListCommit(ctx, db, filter, false)
		require.NoError(t, err, "should be able to list repos")
		checkOutput(ctx, t, iter, expectedInfos)
	})
}

type commitTestCase func(context.Context, *testing.T, *pachsql.DB)

func testCommitDataModelAPI(t *testing.T, testCase commitTestCase) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.State_2_8_0_temp), "should be able to set up tables")
	testCase(ctx, t, db)
}

func commitsMatch(t *testing.T, a, b *pfs.CommitInfo) {
	require.Equal(t, a.Commit.Repo.Name, b.Commit.Repo.Name)
	require.Equal(t, a.Commit.Id, b.Commit.Id)
	if a.Commit.Branch != nil || b.Commit.Branch != nil {
		require.Equal(t, a.Commit.Branch.Name, b.Commit.Branch.Name)
	}
	require.Equal(t, a.Origin.Kind, b.Origin.Kind)
	require.Equal(t, a.Description, b.Description)
	require.Equal(t, a.Started.Seconds, b.Started.Seconds)
	if a.ParentCommit != nil || b.ParentCommit != nil {
		require.Equal(t, a.ParentCommit.Id, b.ParentCommit.Id)
		require.Equal(t, a.ParentCommit.Repo.Name, b.ParentCommit.Repo.Name)
	}
	require.Equal(t, len(a.ChildCommits), len(b.ChildCommits))
	if len(a.ChildCommits) != 0 || len(b.ChildCommits) != 0 {
		childMap := make(map[string]*pfs.Commit)
		for _, commit := range a.ChildCommits {
			childMap[pfsdb.CommitKey(commit)] = commit
		}
		for _, commit := range b.ChildCommits {
			require.Equal(t, commit.Id, childMap[pfsdb.CommitKey(commit)].Id)
			require.Equal(t, commit.Repo.Name, childMap[pfsdb.CommitKey(commit)].Repo.Name)
		}
	}
}

func testCommit(ctx context.Context, t *testing.T, tx *pachsql.Tx, repoName string) *pfs.CommitInfo {
	repoInfo := testRepo(repoName, testRepoType)
	var commitInfo *pfs.CommitInfo
	id := random.String(32)
	commit := &pfs.Commit{Repo: repoInfo.Repo, Branch: &pfs.Branch{Repo: repoInfo.Repo, Name: "master"}, Id: id}
	commitInfo = &pfs.CommitInfo{
		Commit:      commit,
		Description: "fake commit",
		Origin: &pfs.CommitOrigin{
			Kind: pfs.OriginKind_USER,
		},
		Started: timestamppb.New(time.Now()),
	}
	_, err := pfsdb.UpsertRepo(ctx, tx, repoInfo)
	require.NoError(t, err, "should be able to create repo")
	return commitInfo
}

func createBranch(ctx context.Context, t *testing.T, tx *pachsql.Tx, commit *pfs.Commit) {
	branchInfo := &pfs.BranchInfo{Branch: commit.Branch, Head: commit}
	_, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
	require.NoError(t, err, "should be able to create branch")
}
