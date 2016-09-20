package persist

import (
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"

	"github.com/dancannon/gorethink"
)

// Indexes is a collection of indexes for easier initialization
var Indexes = []*index{
	DiffPathIndex,
	DiffPrefixIndex,
	DiffParentIndex,
	DiffClockIndex,
	CommitBranchIndex,
	CommitClockIndex,
	CommitFullClockIndex,
}

// index is a rethinkdb index.
type index struct {
	Name           string
	Table          Table
	CreateFunction func(gorethink.Term) interface{}
	CreateOptions  gorethink.IndexCreateOpts
}

// DiffPathIndex maps a path to diffs that for that path
// Format: [repo, path, clocks]
// Example:
// For the diff: "/foo/bar/buzz", (master, 1)
// We'd have the following index entries:
// ["/foo/bar/buzz", (master, 1)]
var DiffPathIndex = &index{
	Name:  "diffPathIndex",
	Table: diffTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		return []interface{}{row.Field("Repo"), row.Field("Path"), persist.ClockToArray(row.Field("Clock"))}
	},
}

func diffPathIndexKey(repo interface{}, path interface{}, clock interface{}) interface{} {
	return []interface{}{repo, path, clock}
}

// DiffPrefixIndex maps a path to diffs that have the path as prefix // Format: [repo, prefix, clocks]
// Example:
// For the diff: "/foo/bar/buzz", (master, 1)
// We'd have the following index entries:
// ["/", (master, 1)]
// ["/foo", (master, 1)]
// ["/foo/bar", (master, 1)]
var DiffPrefixIndex = &index{
	Name:  "DiffPrefixIndex",
	Table: diffTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		return row.Field("Path").Split("/").DeleteAt(-1).Fold("", func(acc, part gorethink.Term) gorethink.Term {
			return gorethink.Branch(
				acc.Eq("/"),
				acc.Add(part),
				acc.Add("/").Add(part),
			)
		}, gorethink.FoldOpts{
			Emit: func(acc, row, newAcc gorethink.Term) []interface{} {
				return []interface{}{newAcc}
			},
		}).Map(func(path gorethink.Term) interface{} {
			return []interface{}{row.Field("Repo"), path, persist.ClockToArray(row.Field("Clock"))}
		})
	},
	CreateOptions: gorethink.IndexCreateOpts{
		Multi: true,
	},
}

func diffPrefixIndexKey(repo interface{}, path interface{}, clock interface{}) interface{} {
	return []interface{}{repo, path, clock}
}

// DiffParentIndex maps a path to diffs that have the path as direct parent
// Format: [repo, parent, clocks]
// Example:
// For the diff: "/foo/bar/buzz", (master, 1)
// We'd have the following index entries:
// ["/foo/bar", (master, 1)]
var DiffParentIndex = &index{
	Name:  "DiffParentIndex",
	Table: diffTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		parent := row.Field("Path").Split("/").DeleteAt(-1).Fold("", func(acc, part gorethink.Term) gorethink.Term {
			return gorethink.Branch(
				acc.Eq("/"),
				acc.Add(part),
				acc.Add("/").Add(part),
			)
		})
		return []interface{}{row.Field("Repo"), parent, persist.ClockToArray(row.Field("Clock"))}
	},
}

func diffParentIndexKey(repo interface{}, path interface{}, clock interface{}) interface{} {
	return []interface{}{repo, path, clock}
}

// DiffClockIndex maps a clock to diffs
// Format: [repo, branch, clock]
// Example: ["test", "master", 1]
var DiffClockIndex = &index{
	Name:  "DiffClockIndex",
	Table: diffTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		clock := row.Field("Clock")
		return []interface{}{row.Field("Repo"), clock.Field("Branch"), clock.Field("Clock")}
	},
}

func diffClockIndexKey(repo interface{}, branch interface{}, clock interface{}) interface{} {
	return []interface{}{repo, branch, clock}
}

// CommitBranchIndex maps clocks to branches
// Format: repo + branch
// Example:
// A commit that has the clock [(master, 2), (foo, 3)] will be indexed to:
// ["repo", "foo"]
var CommitBranchIndex = &index{
	Name:  "CommitBranchIndex",
	Table: commitTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		lastClock := row.Field("FullClock").Nth(-1)
		return []interface{}{
			row.Field("Repo"),
			lastClock.Field("Branch"),
		}
	},
}

func commitBranchIndexKey(repo interface{}, branch interface{}) interface{} {
	return []interface{}{repo, branch}
}

// CommitClockIndex maps clocks to commits
// Format: repo + head of clocks
// Example:
// A commit that has the clock [(master, 2), (foo, 3)] will be indexed to:
// ["repo", "foo", 3]
var CommitClockIndex = &index{
	Name:  "CommitClockIndex",
	Table: commitTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		lastClock := row.Field("FullClock").Nth(-1)
		return []interface{}{
			row.Field("Repo"),
			lastClock.Field("Branch"),
			lastClock.Field("Clock"),
		}
	},
}

func commitClockIndexKey(repo interface{}, branch interface{}, clock interface{}) interface{} {
	return []interface{}{repo, branch, clock}
}

// CommitFullClockIndex indexes the FullClock of a commit
var CommitFullClockIndex = &index{
	Name:  "CommitFullClockIndex",
	Table: commitTable,
	CreateFunction: func(row gorethink.Term) interface{} {
		return []interface{}{
			row.Field("Repo"),
			persist.FullClockToArray(row.Field("FullClock")),
		}
	},
}
