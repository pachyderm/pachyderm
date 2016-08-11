package persist

import (
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"

	"github.com/dancannon/gorethink"
)

var (
	DiffPathIndex     = NewDiffPathIndex()
	DiffPrefixIndex   = NewDiffPrefixIndex()
	DiffParentIndex   = NewDiffParentIndex()
	DiffCommitIndex   = NewDiffCommitIndex()
	ClockBranchIndex  = NewClockBranchIndex()
	CommitBranchIndex = NewCommitBranchIndex()

	// Collect them all for easier initialization
	Indexes = []Index{
		DiffPathIndex,
		DiffPrefixIndex,
		DiffParentIndex,
		DiffCommitIndex,
		ClockBranchIndex,
		CommitBranchIndex,
	}
)

// An Index is a rethinkdb index.
type Index interface {
	GetName() string
	GetTable() Table
	GetCreateFunction() func(gorethink.Term) interface{}
	GetCreateOptions() gorethink.IndexCreateOpts
}

type index struct {
	Name           string
	Table          Table
	CreateFunction func(gorethink.Term) interface{}
	CreateOptions  gorethink.IndexCreateOpts
}

func (i *index) GetName() string {
	return i.Name
}

func (i *index) GetTable() Table {
	return i.Table
}

func (i *index) GetCreateFunction() func(gorethink.Term) interface{} {
	return i.CreateFunction
}

func (i *index) GetCreateOptions() gorethink.IndexCreateOpts {
	return i.CreateOptions
}

// diffPathIndex maps a path to diffs that for that path
// Format: [repo, path, clocks]
// Example:
// For the diff: delete "/foo/bar/buzz", [(master, 1), (foo, 2)]
// We'd have the following index entries:
// ["/foo/bar/buzz", (master, 1)]
// ["/foo/bar/buzz", (foo, 2)]
type diffPathIndex struct {
	index
}

func NewDiffPathIndex() *diffPathIndex {
	return &diffPathIndex{index{
		Name:  "diffPathIndex",
		Table: diffTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return row.Field("Clocks").Map(func(clock gorethink.Term) interface{} {
				return []interface{}{row.Field("Repo"), row.Field("Path"), persist.ClockToArray(clock)}
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}}
}

func (i *diffPathIndex) Key(repo interface{}, path interface{}, clock interface{}) interface{} {
	return []interface{}{repo, path, clock}
}

// diffPrefixIndex maps a path to diffs that have the path as prefix // Format: [repo, prefix, clocks]
// Example:
// For the diff: "/foo/bar/buzz", [(master, 1), (foo, 2)]
// We'd have the following index entries:
// ["/", (master, 1)]
// ["/foo", (master, 1)]
// ["/foo/bar", (master, 1)]
// ["/", (foo, 2)]
// ["/foo", (foo, 2)]
// ["/foo/bar", (foo, 2)]
type diffPrefixIndex struct {
	index
}

func NewDiffPrefixIndex() *diffPrefixIndex {
	return &diffPrefixIndex{index{
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
			}).ConcatMap(func(path gorethink.Term) gorethink.Term {
				return row.Field("Clocks").Map(func(clock gorethink.Term) interface{} {
					return []interface{}{row.Field("Repo"), path, persist.ClockToArray(clock)}
				})
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}}
}

func (i *diffPrefixIndex) Key(repo interface{}, path interface{}, clock interface{}) interface{} {
	return []interface{}{repo, path, clock}
}

// diffParentIndex maps a path to diffs that have the path as direct parent
// Format: [repo, parent, clocks]
// Example:
// For the diff: "/foo/bar/buzz", [(master, 1), (foo, 2)]
// We'd have the following index entries:
// ["/foo/bar", (master, 1)]
// ["/foo/bar", (foo, 2)]
type diffParentIndex struct {
	index
}

func NewDiffParentIndex() *diffParentIndex {
	return &diffParentIndex{index{
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
			return row.Field("clocks").Map(func(clock gorethink.Term) interface{} {
				return []interface{}{row.Field("Repo"), parent, persist.ClockToArray(clock)}
			})
		},
		CreateOptions: gorethink.IndexCreateOpts{
			Multi: true,
		},
	}}
}

func (i *diffParentIndex) Key(repo interface{}, path interface{}, clock interface{}) interface{} {
	return []interface{}{repo, path, clock}
}

// diffCommitIndex maps a commit ID to diffs
// Format: commit ID
// Example: "vswS3kJkejCnyVJLkHfjbUrSwnSAgHpW"
type diffCommitIndex struct {
	index
}

func NewDiffCommitIndex() *diffCommitIndex {
	return &diffCommitIndex{index{
		Name:  "DiffCommitIndex",
		Table: diffTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return row.Field("CommitID")
		},
	}}
}

type clockBranchIndex struct {
	index
}

func NewClockBranchIndex() *clockBranchIndex {
	return &clockBranchIndex{index{
		Name:  "ClockBranchIndex",
		Table: clockTable,
		CreateFunction: func(row gorethink.Term) interface{} {
			return []interface{}{
				row.Field("Repo"),
				row.Field("Branch"),
			}
		},
	}}
}

// commitBranchIndex maps branch positions to commits
// Format: repo + head of clocks
// Example:
// A commit that has the clock [(master, 2), (foo, 3)] will be indexed to:
// ["repo", "foo", 3]

type commitBranchIndex struct {
	index
}

func (i *commitBranchIndex) Key(repo interface{}, branch interface{}, clock interface{}) interface{} {
	return []interface{}{repo, branch, clock}
}

func NewCommitBranchIndex() *commitBranchIndex {
	return &commitBranchIndex{index{
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
	}}
}
