package pfsdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func compareHead(expected, got *pfs.Commit) bool {
	return expected.Id == got.Id &&
		expected.Repo.Name == got.Repo.Name &&
		expected.Repo.Type == got.Repo.Type &&
		expected.Repo.Project.Name == got.Repo.Project.Name
}

func compareBranch(expected, got *pfs.Branch) bool {
	return expected.Name == got.Name &&
		expected.Repo.Name == got.Repo.Name &&
		expected.Repo.Type == got.Repo.Type &&
		expected.Repo.Project.Name == got.Repo.Project.Name
}

func compareBranchOpts() []cmp.Option {
	return []cmp.Option{
		cmpopts.IgnoreUnexported(pfs.BranchInfo{}),
		cmpopts.SortSlices(func(a, b *pfs.Branch) bool { return a.Key() < b.Key() }), // Note that this is before compareBranch because we need to sort first.
		cmpopts.SortMaps(func(a, b pfsdb.BranchID) bool { return a < b }),
		cmpopts.EquateEmpty(),
		cmp.Comparer(compareBranch),
		cmp.Comparer(compareHead),
	}
}

func newProjectInfo(name string) *pfs.ProjectInfo {
	return &pfs.ProjectInfo{
		Project: &pfs.Project{
			Name: name,
		},
		Description: "test project",
	}
}

func newRepoInfo(project *pfs.Project, name, repoType string) *pfs.RepoInfo {
	return &pfs.RepoInfo{
		Repo: &pfs.Repo{
			Project: project,
			Name:    name,
			Type:    repoType,
		},
		Description: "test repo",
	}
}

func newCommitInfo(repo *pfs.Repo, id string, parent *pfs.Commit) *pfs.CommitInfo {
	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Repo:   repo,
			Id:     id,
			Branch: &pfs.Branch{},
		},
		Description:  "test commit",
		ParentCommit: parent,
		Origin:       &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
		Started:      timestamppb.New(time.Now()),
	}
}

func createProject(t *testing.T, ctx context.Context, tx *pachsql.Tx, projectInfo *pfs.ProjectInfo) {
	t.Helper()
	require.NoError(t, coredb.UpsertProject(ctx, tx, projectInfo))
}

func createRepoInfoWithID(t *testing.T, ctx context.Context, tx *pachsql.Tx, repoInfo *pfs.RepoInfo) *pfsdb.RepoPair {
	t.Helper()
	createProject(t, ctx, tx, newProjectInfo(repoInfo.Repo.Project.Name))
	id, err := pfsdb.UpsertRepo(ctx, tx, repoInfo)
	require.NoError(t, err)
	return &pfsdb.RepoPair{ID: id, RepoInfo: repoInfo}
}

func createCreateInfoWithID(t *testing.T, ctx context.Context, tx *pachsql.Tx, commitInfo *pfs.CommitInfo) *pfsdb.CommitWithID {
	t.Helper()
	createRepoInfoWithID(t, ctx, tx, newRepoInfo(commitInfo.Commit.Repo.Project, commitInfo.Commit.Repo.Name, commitInfo.Commit.Repo.Type))
	commitID, err := pfsdb.CreateCommit(ctx, tx, commitInfo)
	require.NoError(t, err)
	return &pfsdb.CommitWithID{ID: commitID, CommitInfo: commitInfo}
}

func TestBranchUpsert(t *testing.T) {
	t.Parallel()
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			repoInfo := newRepoInfo(&pfs.Project{Name: "project1"}, "repo1", pfs.UserRepoType)
			commitInfoWithID1 := createCreateInfoWithID(t, ctx, tx, newCommitInfo(repoInfo.Repo, random.String(32), nil))
			branchInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoInfo.Repo,
					Name: "master",
				},
				Head: commitInfoWithID1.CommitInfo.Commit,
			}
			id, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
			require.NoError(t, err)
			gotBranchInfo, err := pfsdb.GetBranchInfo(ctx, tx, id)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo, gotBranchInfo, compareBranchOpts()...))
			gotBranchByName, err := pfsdb.GetBranchInfoByName(ctx, tx, branchInfo.Branch.Repo.Project.Name, branchInfo.Branch.Repo.Name, branchInfo.Branch.Repo.Type, branchInfo.Branch.Name)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo, gotBranchByName, compareBranchOpts()...))

			// Update branch to point to second commit
			commitInfoWithID2 := createCreateInfoWithID(t, ctx, tx, newCommitInfo(repoInfo.Repo, random.String(32), commitInfoWithID1.CommitInfo.Commit))
			branchInfo.Head = commitInfoWithID2.CommitInfo.Commit
			id2, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
			require.NoError(t, err)
			require.Equal(t, id, id2, "UpsertBranch should keep id stable")
			gotBranchInfo2, err := pfsdb.GetBranchInfo(ctx, tx, id2)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchInfo, gotBranchInfo2, compareBranchOpts()...))
		})
	})
}

func TestBranchProvenance(t *testing.T) {
	t.Parallel()
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			// Setup dependencies
			repoAInfo := newRepoInfo(&pfs.Project{Name: pfs.DefaultProjectName}, "A", pfs.UserRepoType)
			repoBInfo := newRepoInfo(&pfs.Project{Name: pfs.DefaultProjectName}, "B", pfs.UserRepoType)
			repoCInfo := newRepoInfo(&pfs.Project{Name: pfs.DefaultProjectName}, "C", pfs.UserRepoType)
			commitAInfo := newCommitInfo(repoAInfo.Repo, random.String(32), nil)
			commitBInfo := newCommitInfo(repoBInfo.Repo, random.String(32), nil)
			commitCInfo := newCommitInfo(repoCInfo.Repo, random.String(32), nil)
			for _, commitInfo := range []*pfs.CommitInfo{commitAInfo, commitBInfo, commitCInfo} {
				createCreateInfoWithID(t, ctx, tx, commitInfo)
			}
			// Create 3 branches, one for each repo, pointing to the corresponding commit
			branchAInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoAInfo.Repo,
					Name: "master",
				},
				Head: commitAInfo.Commit,
			}
			branchBInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoBInfo.Repo,
					Name: "master",
				},
				Head: commitBInfo.Commit,
			}
			branchCInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoCInfo.Repo,
					Name: "master",
				},
				Head: commitCInfo.Commit,
			}
			// Provenance info: A <- B <- C, and A <- C
			branchAInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch, branchCInfo.Branch}
			branchBInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch}
			branchBInfo.Provenance = []*pfs.Branch{branchAInfo.Branch}
			branchBInfo.Subvenance = []*pfs.Branch{branchCInfo.Branch}
			branchCInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch, branchBInfo.Branch}
			branchCInfo.Provenance = []*pfs.Branch{branchBInfo.Branch, branchAInfo.Branch}
			// Create all branches, and provenance relationships
			allBranches := make(map[pfsdb.BranchID]*pfs.BranchInfo)
			for _, branchInfo := range []*pfs.BranchInfo{branchAInfo, branchBInfo, branchCInfo} {
				id, err := pfsdb.UpsertBranch(ctx, tx, branchInfo) // implicitly creates prov relationships
				require.NoError(t, err)
				allBranches[id] = branchInfo
			}
			// Verify direct provenance, full provenance, and full subvenance relationships
			for id, branchInfo := range allBranches {
				gotDirectProv, err := pfsdb.GetDirectBranchProvenance(ctx, tx, id)
				require.NoError(t, err)
				require.True(t, cmp.Equal(branchInfo.DirectProvenance, gotDirectProv, compareBranchOpts()...))
				gotProv, err := pfsdb.GetBranchProvenance(ctx, tx, id)
				require.NoError(t, err)
				require.True(t, cmp.Equal(branchInfo.Provenance, gotProv, compareBranchOpts()...))
				gotSubv, err := pfsdb.GetBranchSubvenance(ctx, tx, id)
				require.NoError(t, err)
				require.True(t, cmp.Equal(branchInfo.Subvenance, gotSubv, compareBranchOpts()...))
			}
			// Update provenance DAG to A <- B -> C, to test adding and deleting prov relationships
			branchAInfo.DirectProvenance = nil
			branchAInfo.Provenance = nil
			branchAInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch}
			branchBInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch, branchCInfo.Branch}
			branchBInfo.Provenance = []*pfs.Branch{branchAInfo.Branch, branchCInfo.Branch}
			branchBInfo.Subvenance = nil
			branchCInfo.DirectProvenance = nil
			branchCInfo.Provenance = nil
			branchCInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch}
			for id, branchInfo := range allBranches {
				gotID, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
				require.NoError(t, err)
				require.Equal(t, id, gotID, "UpsertBranch should keep id stable")
			}
			for id, branchInfo := range allBranches {
				gotDirectProv, err := pfsdb.GetDirectBranchProvenance(ctx, tx, id)
				require.NoError(t, err)
				require.True(t, cmp.Equal(branchInfo.DirectProvenance, gotDirectProv, compareBranchOpts()...))
				gotProv, err := pfsdb.GetBranchProvenance(ctx, tx, id)
				require.NoError(t, err)
				require.True(t, cmp.Equal(branchInfo.Provenance, gotProv, compareBranchOpts()...))
				gotSubv, err := pfsdb.GetBranchSubvenance(ctx, tx, id)
				require.NoError(t, err)
				require.True(t, cmp.Equal(branchInfo.Subvenance, gotSubv, compareBranchOpts()...))
			}
		})
	})
}

func TestBranchIterator(t *testing.T) {
	t.Parallel()
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		allBranches := make(map[pfsdb.BranchID]*pfs.BranchInfo)
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			// Create 2^8-1=255 branches
			headCommitInfo := newCommitInfo(&pfs.Repo{Project: &pfs.Project{Name: "test-project"}, Name: testutil.UniqueString("test-repo"), Type: pfs.UserRepoType}, random.String(32), nil)
			rootBranchInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: headCommitInfo.Commit.Repo,
					Name: "master",
				},
				Head: headCommitInfo.Commit,
			}

			currentLevel := []*pfs.BranchInfo{rootBranchInfo}
			for i := 0; i < 8; i++ {
				var newLevel []*pfs.BranchInfo
				for _, parent := range currentLevel {
					// create a commits and branches
					createCreateInfoWithID(t, ctx, tx, newCommitInfo(parent.Head.Repo, parent.Head.Id, nil))
					id, err := pfsdb.UpsertBranch(ctx, tx, parent)
					require.NoError(t, err)
					allBranches[id] = parent
					// Create 2 child for each branch in the current level
					for j := 0; j < 2; j++ {
						head := newCommitInfo(&pfs.Repo{Project: &pfs.Project{Name: "test-project"}, Name: testutil.UniqueString("test-repo"), Type: pfs.UserRepoType}, random.String(32), nil)
						child := &pfs.BranchInfo{
							Branch: &pfs.Branch{
								Repo: head.Commit.Repo,
								Name: "master",
							},
							Head: head.Commit,
						}
						newLevel = append(newLevel, child)
					}
				}
				currentLevel = newLevel
			}
		})
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			// List all branches
			branchIterator, err := pfsdb.NewBranchIterator(ctx, tx, 0 /* startPage */, 10 /* pageSize */, nil /* filter */)
			require.NoError(t, err)
			gotAllBranches := make(map[pfsdb.BranchID]*pfs.BranchInfo)
			require.NoError(t, stream.ForEach[pfsdb.BranchInfoWithID](ctx, branchIterator, func(branchInfoWithID pfsdb.BranchInfoWithID) error {
				gotAllBranches[branchInfoWithID.ID] = branchInfoWithID.BranchInfo
				require.Equal(t, allBranches[branchInfoWithID.ID].Branch.Key(), branchInfoWithID.BranchInfo.Branch.Key())
				require.Equal(t, allBranches[branchInfoWithID.ID].Head.Key(), branchInfoWithID.BranchInfo.Head.Key())
				return nil
			}))
			// Filter on a set of repos
			expectedRepoNames := []string{allBranches[1].Branch.Repo.Name}
			branchIterator, err = pfsdb.NewBranchIterator(ctx, tx, 0 /* startPage */, 10 /* pageSize */, allBranches[1].Branch, pfsdb.OrderByBranchColumn{Column: pfsdb.BranchColumnCreatedAt, Order: pfsdb.SortOrderAsc})
			require.NoError(t, err)
			var gotRepoNames []string
			require.NoError(t, stream.ForEach[pfsdb.BranchInfoWithID](ctx, branchIterator, func(branchInfoWithID pfsdb.BranchInfoWithID) error {
				gotRepoNames = append(gotRepoNames, branchInfoWithID.BranchInfo.Branch.Repo.Name)
				return nil
			}))
			require.Equal(t, len(expectedRepoNames), len(gotRepoNames))
			require.ElementsEqual(t, expectedRepoNames, gotRepoNames)
		})
	})
}

func TestBranchDelete(t *testing.T) {
	t.Parallel()
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			// Setup dependencies
			repoInfo := newRepoInfo(&pfs.Project{Name: pfs.DefaultProjectName}, "A", pfs.UserRepoType)
			commitAInfo := newCommitInfo(repoInfo.Repo, random.String(32), nil)
			commitBInfo := newCommitInfo(repoInfo.Repo, random.String(32), nil)
			commitCInfo := newCommitInfo(repoInfo.Repo, random.String(32), nil)
			for _, commitInfo := range []*pfs.CommitInfo{commitAInfo, commitBInfo, commitCInfo} {
				createCreateInfoWithID(t, ctx, tx, commitInfo)
			}
			// Create 3 branches, one for each repo, pointing to the corresponding commit
			branchAInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoInfo.Repo,
					Name: "branchA",
				},
				Head: commitAInfo.Commit,
			}
			branchBInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoInfo.Repo,
					Name: "branchB",
				},
				Head: commitBInfo.Commit,
			}
			branchCInfo := &pfs.BranchInfo{
				Branch: &pfs.Branch{
					Repo: repoInfo.Repo,
					Name: "branchC",
				},
				Head: commitCInfo.Commit,
			}
			// Provenance info: A <- B <- C, and A <- C
			branchAInfo.Subvenance = []*pfs.Branch{branchBInfo.Branch, branchCInfo.Branch}
			branchBInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch}
			branchBInfo.Provenance = []*pfs.Branch{branchAInfo.Branch}
			branchBInfo.Subvenance = []*pfs.Branch{branchCInfo.Branch}
			branchCInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch, branchBInfo.Branch}
			branchCInfo.Provenance = []*pfs.Branch{branchBInfo.Branch, branchAInfo.Branch}
			// Add a branch trigger to re-point branch C to B
			branchCInfo.Trigger = &pfs.Trigger{Branch: "branchB", CronSpec: "* * * * *"}
			for _, branchInfo := range []*pfs.BranchInfo{branchAInfo, branchBInfo, branchCInfo} {
				_, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
				require.NoError(t, err)
			}
			branchBID, err := pfsdb.GetBranchID(ctx, tx, branchBInfo.Branch)
			require.NoError(t, err)
			require.NoError(t, pfsdb.DeleteBranch(ctx, tx, branchBID))
			_, err = pfsdb.GetBranchInfo(ctx, tx, branchBID)
			require.ErrorContains(t, err, "sql: no rows in result set")
			// Verify that BranchA no longer has BranchB in its subvenance
			branchAInfo.Subvenance = []*pfs.Branch{branchCInfo.Branch}
			branchAID, err := pfsdb.GetBranchID(ctx, tx, branchAInfo.Branch)
			require.NoError(t, err)
			gotBranchAInfo, err := pfsdb.GetBranchInfo(ctx, tx, branchAID)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchAInfo, gotBranchAInfo, compareBranchOpts()...))
			// Verify BranchC no longer has BranchB in its provenance, nor does it have the trigger
			branchCInfo.DirectProvenance = []*pfs.Branch{branchAInfo.Branch}
			branchCInfo.Provenance = []*pfs.Branch{branchAInfo.Branch}
			branchCInfo.Trigger = nil
			branchCID, err := pfsdb.GetBranchID(ctx, tx, branchCInfo.Branch)
			require.NoError(t, err)
			gotBranchCInfo, err := pfsdb.GetBranchInfo(ctx, tx, branchCID)
			require.NoError(t, err)
			require.True(t, cmp.Equal(branchCInfo, gotBranchCInfo, compareBranchOpts()...))
		})
	})
}

func TestBranchTrigger(t *testing.T) {
	t.Parallel()
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		var masterBranchID, stagingBranchID pfsdb.BranchID
		// Create two branches, master and staging in the same repo.
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			var err error
			repoInfo := newRepoInfo(&pfs.Project{Name: "project1"}, "repo1", pfs.UserRepoType)
			commit1 := createCreateInfoWithID(t, ctx, tx, newCommitInfo(repoInfo.Repo, random.String(32), nil)).CommitInfo.Commit
			masterBranchInfo := &pfs.BranchInfo{Branch: &pfs.Branch{Repo: repoInfo.Repo, Name: "master"}, Head: commit1}
			masterBranchID, err = pfsdb.UpsertBranch(ctx, tx, masterBranchInfo)
			require.NoError(t, err)
			commit2 := createCreateInfoWithID(t, ctx, tx, newCommitInfo(repoInfo.Repo, random.String(32), nil)).CommitInfo.Commit
			stagingBranchInfo := &pfs.BranchInfo{Branch: &pfs.Branch{Repo: repoInfo.Repo, Name: "staging"}, Head: commit2}
			stagingBranchID, err = pfsdb.UpsertBranch(ctx, tx, stagingBranchInfo)
			require.NoError(t, err)
		})
		// Create the branch trigger that re-points master to staging.
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			trigger := &pfs.Trigger{
				Branch:        "staging",
				CronSpec:      "* * * * *",
				RateLimitSpec: "",
				Size:          "100M",
				Commits:       10,
				All:           true,
			}
			require.NoError(t, pfsdb.UpsertBranchTrigger(ctx, tx, masterBranchID, stagingBranchID, trigger))
			gotTrigger, err := pfsdb.GetBranchTrigger(ctx, tx, masterBranchID)
			require.NoError(t, err)
			require.Equal(t, trigger, gotTrigger)
			// Also get the trigger from GetBranchInfo
			gotMasterBranchInfo, err := pfsdb.GetBranchInfo(ctx, tx, masterBranchID)
			require.NoError(t, err)
			require.Equal(t, trigger, gotMasterBranchInfo.Trigger)
			// Update the trigger through UpsertBranchTrigger
			trigger.CronSpec = "0 * * * *"
			trigger.All = false
			require.NoError(t, pfsdb.UpsertBranchTrigger(ctx, tx, masterBranchID, stagingBranchID, trigger))
			gotMasterBranchInfo, err = pfsdb.GetBranchInfo(ctx, tx, masterBranchID)
			require.NoError(t, err)
			require.Equal(t, trigger, gotMasterBranchInfo.Trigger)
			// Delete branch trigger, and try to get it back via GetBranchInfo
			require.NoError(t, pfsdb.DeleteBranchTrigger(ctx, tx, masterBranchID))
			gotMasterBranchInfo, err = pfsdb.GetBranchInfo(ctx, tx, masterBranchID)
			require.NoError(t, err)
			require.Nil(t, gotMasterBranchInfo.Trigger)
			// staging branch shouldn't get a trigger
			gotStagingBranchInfo, err := pfsdb.GetBranchInfo(ctx, tx, stagingBranchID)
			require.NoError(t, err)
			require.Nil(t, gotStagingBranchInfo.Trigger)
			// Attempt to create trigger with nonexistent branch via UpsertBranch
			gotMasterBranchInfo.Trigger = &pfs.Trigger{Branch: "nonexistent"}
			_, err = pfsdb.UpsertBranch(ctx, tx, gotMasterBranchInfo)
			require.ErrorContains(t, err, "no rows in result set")
		})
	})
}
