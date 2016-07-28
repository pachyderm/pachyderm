package persist

import (
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"

	"github.com/dancannon/gorethink"
)

// An Index is a rethinkdb index.
type Index struct {
	Name           string
	Table          Table
	CreateFunction func(gorethink.Term) interface{}
	CreateOptions  gorethink.IndexCreateOpts
}

var (
	// diffPathIndex maps a path to diffs that for that path
	// Format: [repo, bool, path, clocks]
	// The bool specifies whether it's a delete.
	// Example:
	// For the diff: delete "/foo/bar/buzz", [[(master, 1)], [(master, 0), (foo, 2)]]
	// We'd have the following index entries:
	// ["/foo/bar/buzz", true, [(master, 1)]]
	// ["/foo/bar/buzz", true, [(master, 0), (foo, 2)]]
	diffPathIndex = Index{
		Name:  "diffPathIndex",
		Table: diffTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return row.Field("BranchClocks").Map(func(branchClock gorethink.Term) interface{} {
				return []interface{}{row.Field("Repo"), row.Field("Delete"), row.Field("Path"), persist.BranchClockToArray(branchClock)}
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}

	// diffPrefixIndex maps a path to diffs that have the path as prefix // Format: [repo, prefix, clocks]
	// Example:
	// For the diff: "/foo/bar/buzz", [[(master, 1)], [(master, 0), (foo, 2)]]
	// We'd have the following index entries:
	// ["/foo", [(master, 1)]]
	// ["/foo/bar", [(master, 1)]]
	// ["/foo", [(master, 0), (foo, 2)]]
	// ["/foo/bar", [(master, 0), (foo, 2)]]
	diffPrefixIndex = Index{
		Name:  "DiffPrefixIndex",
		Table: diffTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return row.Field("Path").Split("/").DeleteAt(-1).Fold("", func(acc, part gorethink.Term) gorethink.Term {
				return acc.Add("/").Add(part)
			}, gorethink.FoldOpts{
				Emit: func(acc, row, newAcc gorethink.Term) gorethink.Term {
					return newAcc
				},
			}).ConcatMap(func(path gorethink.Term) gorethink.Term {
				return row.Field("BranchClocks").Map(func(branchClock gorethink.Term) interface{} {
					return []interface{}{row.Field("Repo"), path, persist.BranchClockToArray(branchClock)}
				})
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}

	// diffParentIndex maps a path to diffs that have the path as direct parent
	// Format: [repo, parent, clocks]
	// Example:
	// For the diff: "/foo/bar/buzz", [[(master, 1)], [(master, 0), (foo, 2)]]
	// We'd have the following index entries:
	// ["/foo/bar", [(master, 1)]]
	// ["/foo/bar", [(master, 0), (foo, 2)]]
	diffParentIndex = Index{
		Name:  "DiffParentIndex",
		Table: diffTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			parent := row.Field("Path").Split("/").DeleteAt(-1).Fold("", func(acc, part gorethink.Term) gorethink.Term {
				return acc.Add("/").Add(part)
			})
			return row.Field("BranchClocks").Map(func(branchClock gorethink.Term) interface{} {
				return []interface{}{row.Field("Repo"), parent, persist.BranchClockToArray(branchClock)}
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}

	// diffCommitIndex maps a commit ID to diffs
	// Format: commit ID
	// Example: "vswS3kJkejCnyVJLkHfjbUrSwnSAgHpW"
	diffCommitIndex = Index{
		Name:  "DiffCommitIndex",
		Table: diffTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return row.Field("CommitID")
		},
	}

	clockBranchIndex = Index{
		Name:  "ClockBranchIndex",
		Table: clockTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return []interface{}{
				row.Field("Repo"),
				row.Field("Branch"),
			}
		},
	}

	// commitBranchIndex maps branch positions to commits
	// Format: repo + head of clocks
	// Example:
	// A commit that has the clock [(master, 2), (foo, 3)] will be indexed to:
	// ["repo", "foo", 3]
	commitBranchIndex = Index{
		Name:  "CommitBranchIndex",
		Table: commitTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return row.Field("BranchClocks").Map(func(branchClock gorethink.Term) interface{} {
				lastClock := branchClock.Field("Clocks").Nth(-1)
				return []interface{}{
					row.Field("Repo"),
					lastClock.Field("Branch"),
					lastClock.Field("Clock"),
				}
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}

	// Collect them all for easier initialization
	Indexes = []Index{
		diffPathIndex,
		diffPrefixIndex,
		diffParentIndex,
		diffCommitIndex,
		clockBranchIndex,
		commitBranchIndex,
	}
)
