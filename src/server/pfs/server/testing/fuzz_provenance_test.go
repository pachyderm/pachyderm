package testing

import (
	"fmt"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"math/rand"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

const (
	inputRepo          = iota // create a new input repo
	inputBranch               // create a new branch on an existing input repo
	deleteInputBranch         // delete an input branch
	commit                    // commit to an input branch
	squashCommitSet           // squash a commitset from an input branch
	outputRepo                // create a new output repo, with master branch subscribed to random other branches
	outputBranch              // create a new output branch on an existing output repo
	deleteOutputBranch        // delete an output branch
)

func TestFuzzProvenance(t *testing.T) {
	pachClient := pachd.NewTestPachd(t)

	seed := time.Now().UnixNano()
	t.Log("Random seed is", seed)
	r := rand.New(rand.NewSource(seed))

	_, err := pachClient.PfsAPIClient.DeleteAll(pachClient.Ctx(), &emptypb.Empty{})
	require.NoError(t, err)
	nOps := 300
	opShares := []int{
		1, // inputRepo
		1, // inputBranch
		1, // deleteInputBranch
		5, // commit
		3, // squashCommitSet
		1, // outputRepo
		2, // outputBranch
		1, // deleteOutputBranch
	}
	total := 0
	for _, v := range opShares {
		total += v
	}
	var (
		inputRepos     []string
		inputBranches  []*pfs.Branch
		commits        []*pfs.Commit
		outputBranches []*pfs.Branch
	)
OpLoop:
	for i := 0; i < nOps; i++ {
		roll := r.Intn(total)
		if i < 0 {
			roll = inputRepo
		}
		var op int
		for _op, v := range opShares {
			roll -= v
			if roll < 0 {
				op = _op
				break
			}
		}
		switch op {
		case inputRepo:
			repo := tu.UniqueString("repo")
			require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
			inputRepos = append(inputRepos, repo)
			require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", nil))
			inputBranches = append(inputBranches, client.NewBranch(pfs.DefaultProjectName, repo, "master"))
		case inputBranch:
			if len(inputRepos) == 0 {
				continue OpLoop
			}
			repo := inputRepos[r.Intn(len(inputRepos))]
			branch := tu.UniqueString("branch")
			require.NoError(t, pachClient.CreateBranch(pfs.DefaultProjectName, repo, branch, "", "", nil))
			inputBranches = append(inputBranches, client.NewBranch(pfs.DefaultProjectName, repo, branch))
		case deleteInputBranch:
			if len(inputBranches) == 0 {
				continue OpLoop
			}
			i := r.Intn(len(inputBranches))
			branch := inputBranches[i]
			err = pachClient.DeleteBranch(pfs.DefaultProjectName, branch.Repo.Name, branch.Name, false)
			// don't fail if the error was just that it couldn't delete the branch without breaking subvenance
			if err != nil && !strings.Contains(err.Error(), fmt.Sprintf("branch %q cannot be deleted because it's in the direct provenance of", branch)) {
				require.NoError(t, err)
			}
			inputBranches = append(inputBranches[:i], inputBranches[i+1:]...)
		case commit:
			if len(inputBranches) == 0 {
				continue OpLoop
			}
			branch := inputBranches[r.Intn(len(inputBranches))]
			commit, err := pachClient.StartCommit(pfs.DefaultProjectName, branch.Repo.Name, branch.Name)
			require.NoError(t, err)
			require.NoError(t, finishCommit(pachClient, branch.Repo.Name, branch.Name, commit.Id))
			// find and finish all commits in output branches, too
			infos, err := pachClient.InspectCommitSet(commit.Id)
			require.NoError(t, err)
			for _, info := range infos {
				if info.Commit.Id != commit.Id {
					continue
				}
				require.NoError(t, finishCommit(pachClient,
					info.Commit.Repo.Name, "", commit.Id))
			}
			commits = append(commits, commit)
		case squashCommitSet:
			if len(commits) == 0 {
				continue OpLoop
			}
			i := r.Intn(len(commits))
			commit := commits[i]

			err := pachClient.SquashCommitSet(commit.Id)
			if pfsserver.IsSquashWithoutChildrenErr(err) {
				err = pachClient.DropCommitSet(commit.Id)
				if pfsserver.IsDropWithChildrenErr(err) {
					// The commitset cannot be squashed or dropped as some commits have children and some commits don't
					continue
				}
			} else if pfsserver.IsSquashWithSuvenanceErr(err) {
				// TODO(acohen4): destructure error and successfully squash all the dependent commit sets
				continue
			}
			require.NoError(t, err)
			commits = append(commits[:i], commits[i+1:]...)
			ris, err := pachClient.ListRepo()
			require.NoError(t, err)
			for _, ri := range ris {
				bis, err := pachClient.ListBranch(pfs.DefaultProjectName, ri.Repo.Name)
				require.NoError(t, err)
				for _, bi := range bis {
					branch := bi.Branch
					info, err := pachClient.InspectCommit(pfs.DefaultProjectName, branch.Repo.Name, branch.Name, "")
					require.NoError(t, err)
					if info.Finishing == nil {
						require.NoError(t, finishCommit(pachClient, branch.Repo.Name, branch.Name, ""))
					}
				}
			}
		case outputRepo:
			if len(inputBranches) == 0 {
				continue OpLoop
			}
			repo := tu.UniqueString("out")
			require.NoError(t, pachClient.CreateRepo(pfs.DefaultProjectName, repo))
			// outputRepos = append(outputRepos, repo)
			var provBranches []*pfs.Branch
			for num, i := range r.Perm(len(inputBranches))[:r.Intn(len(inputBranches))] {
				provBranches = append(provBranches, inputBranches[i])
				if num > 1 {
					break
				}
			}
			err = pachClient.CreateBranch(pfs.DefaultProjectName, repo, "master", "", "", provBranches)
			if err != nil {
				if pfsserver.IsInvalidBranchStructureErr(err) || strings.Contains(err.Error(), "cannot be in the provenance of its own branch") {
					continue
				}
				require.NoError(t, err)
			} else {
				outputBranches = append(outputBranches, client.NewBranch(pfs.DefaultProjectName, repo, "master"))
				if len(provBranches) > 0 {
					require.NoError(t, finishCommit(pachClient, repo, "master", ""))
				}
			}
		case deleteOutputBranch:
			if len(outputBranches) == 0 {
				continue OpLoop
			}
			i := r.Intn(len(outputBranches))
			branch := outputBranches[i]
			err = pachClient.DeleteBranch(pfs.DefaultProjectName, branch.Repo.Name, branch.Name, false)
			// don't fail if the error was just that it couldn't delete the branch without breaking subvenance
			outputBranches = append(outputBranches[:i], outputBranches[i+1:]...)
			if err != nil && !strings.Contains(err.Error(), "break") {
				require.NoError(t, err)
			}
		}
		require.NoError(t, pachClient.FsckFastExit())
	}
	// make sure we can delete at the end
	_, err = pachClient.PfsAPIClient.DeleteAll(pachClient.Ctx(), &emptypb.Empty{})
	require.NoError(t, err)
}
