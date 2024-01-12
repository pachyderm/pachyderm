package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "totals" {
		totals()
	} else {
		commits()
	}
}

type commitCol struct {
	CommitSet string `json:"idx_commitset"`
	Proto     []byte `json:"proto"`
}

func commits() {
	f, err := os.Open("/Users/alon/Downloads/database_2/tables/collections/commits.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var cols []commitCol
	if err := json.Unmarshal([]byte(byteValue), &cols); err != nil {
		panic(err)
	}
	var cis []*pfs.CommitInfo
	var aliases []*pfs.Repo
	var autos []*pfs.Repo
	var users []*pfs.Repo
	for _, col := range cols {
		if col.CommitSet != "5175829a7e064487a7dda7044198665d" {
			continue
		}
		ci := &pfs.CommitInfo{}
		if err := proto.Unmarshal(col.Proto, ci); err != nil {
			panic(err)
		}
		cis = append(cis, ci)
		if ci.Origin.Kind == 4 {
			aliases = append(aliases, ci.Commit.Branch.Repo)
		} else if ci.Origin.Kind == pfs.OriginKind_AUTO {
			autos = append(autos, ci.Commit.Branch.Repo)
		} else if ci.Origin.Kind == pfs.OriginKind_USER {
			users = append(users, ci.Commit.Branch.Repo)
		}
	}
	fmt.Printf("Aliases:::: %v; count:: %v\n", aliases, len(aliases))
	fmt.Printf("Autos:::: %v; count:: %v\n", autos, len(autos))
	fmt.Printf("Users:::: %v; count:: %v\n", users, len(users))
	if err := validateExistingDAGs(cis); err != nil {
		panic(err)
	}

}

func validateExistingDAGs(cis []*pfs.CommitInfo) error {
	// group duplicate commits by branchless key
	duplicates := make(map[string]map[string]*pfs.CommitInfo) // branchless commit key -> { old commit key ->  commit info }
	for _, ci := range cis {
		if _, ok := duplicates[commitBranchlessKey(ci.Commit)]; !ok {
			duplicates[commitBranchlessKey(ci.Commit)] = make(map[string]*pfs.CommitInfo)
		}
		duplicates[commitBranchlessKey(ci.Commit)][oldCommitKey(ci.Commit)] = ci
	}
	// the only duplicate commits we allow and handle in the migration are two commits with a parent/child relationship
	// for any set of commits with the same ID on a repo, we expect the following:
	// 1. all commits has a non-nil Parent.
	// 2. exactly one commit is not an ALIAS commit
	// 3. the commits are in an ancestry chain
	//
	// TODO(provenance): this is all quite complicated and potentially needs to be revisted
	var badCommitSets []string
	for _, dups := range duplicates {
		if len(dups) <= 1 {
			continue
		}
		seen := make(map[string]struct{})
		aliasCount := 0
		var commitSet string
		for _, d := range dups {
			// before 2.6, pfs.OriginKind_ALIAS = 4
			if d.Origin.Kind == 4 {
				aliasCount++
			}
			commitSet = d.Commit.ID
			if d.ParentCommit != nil {
				if _, ok := dups[oldCommitKey(d.ParentCommit)]; ok {
					seen[oldCommitKey(d.Commit)] = struct{}{}
				}
			}
		}
		if len(dups)-len(seen) > 1 || len(dups)-aliasCount != 1 {
			if len(dups)-aliasCount != 1 {
				fmt.Printf("EEEPAAAAA :: %v\n", len(dups)-aliasCount)
			}
			var cs []string
			for _, c := range dups {
				cs = append(cs, oldCommitKey(c.Commit))
			}
			fmt.Printf("DUPS::: %v\n", cs)
			badCommitSets = append(badCommitSets, commitSet)
		}
	}
	if badCommitSets != nil {
		return errors.Errorf("invalid commit sets detected: %v", badCommitSets)
	}
	return nil
}

func repoKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name + "." + repo.Type
}

func oldCommitKey(commit *pfs.Commit) string {
	return branchKey(commit.Branch) + "=" + commit.ID
}

func commitBranchlessKey(commit *pfs.Commit) string {
	return repoKey(commit.Branch.Repo) + "@" + commit.ID
}

func branchKey(branch *pfs.Branch) string {
	return repoKey(branch.Repo) + "@" + branch.Name
}

type commit struct {
	CommitId  string `json:"commit_id"`
	FileSetId string `json:"fileset_id"`
}

func totals() {
	f, err := os.Open("/Users/alon/Downloads/database/tables/pfs/commit_totals.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var cs []commit
	if err := json.Unmarshal([]byte(byteValue), &cs); err != nil {
		panic(err)
	}
	agg := make(map[string][]string)
	for _, c := range cs {
		k := stripBranch(c.CommitId)
		if _, ok := agg[k]; !ok {
			agg[k] = make([]string, 0)
		}
		agg[k] = append(agg[k], c.CommitId)
	}
	for k, v := range agg {
		if len(v) == 1 {
			delete(agg, k)
		}
	}
	fmt.Println(agg)
}

func stripBranch(c string) string {
	branch := regexp.MustCompile(`@[A-Za-z0-9]+=`)
	return branch.ReplaceAllString(c, "@")
}
