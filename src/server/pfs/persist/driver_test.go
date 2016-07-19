package persist

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"

	"github.com/dancannon/gorethink"
	"go.pedge.io/pb/go/google/protobuf"
)

const (
	RethinkAddress = "localhost:28015"
	RethinkTestDB  = "pachyderm_test"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if err := InitDB(RethinkAddress, RethinkTestDB); err != nil {
		panic(err)
	}
	code := m.Run()
	if err := RemoveDB(RethinkAddress, RethinkTestDB); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func timestampNow() *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{Seconds: time.Now().Unix()}
}

func persistCommitToPFSCommit(rawCommit *Commit) *pfs.Commit {
	return &pfs.Commit{
		Repo: &pfs.Repo{
			Name: rawCommit.Repo,
		},
		ID: rawCommit.ID,
	}
}

/*

TODO, write these cases:

- check uniqueness -- see if creating branch w same id results in rethink error
*/

func TestStartCommit(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"master",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err := gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)
	defer func() {
		require.NoError(t, cursor.Close())
	}()

	rawCommit := &Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0])

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)
}

func TestStartCommitJustByBranch(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"master",
		timestampNow(),
		nil,
		nil,
	)
	require.NoError(t, err)

	cursor, err := gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)
	defer func() {
		require.NoError(t, cursor.Close())
	}()

	rawCommit := &Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0])

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

	commitID = uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"master",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err = gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)

	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // Has 2 commits on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 1}, rawCommit.BranchClocks[0].Clocks[0])

	commit = persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

}

func TestStartCommitSpecifyParentAndBranch(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"master",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err := gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)
	defer func() {
		require.NoError(t, cursor.Close())
	}()

	rawCommit := &Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0].Clock)

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

	commit2ID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commit2ID,
		commitID,
		"master",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err = gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)

	rawCommit2 := &Commit{}
	cursor.Next(rawCommit2)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit2)

	require.Equal(t, 1, len(rawCommit2.BranchClocks))          // Only belongs to one branch
	require.Equal(t, 2, len(rawCommit.BranchClocks[0].Clocks)) // Has 2 commits on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[0])
	require.Equal(t, &Clock{Branch: "master", Clock: 1}, rawCommit2.BranchClocks[0].Clocks[1])

	commit2 := persistCommitToPFSCommit(rawCommit2)
	err = d.FinishCommit(commit2, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

}

func TestStartCommitSpecifyParentAndNewBranch(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"master",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err := gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)
	defer func() {
		require.NoError(t, cursor.Close())
	}()

	rawCommit := &Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0].Clock)

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

	commit2ID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commit2ID,
		commitID,
		"foo",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err = gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)

	rawCommit2 := &Commit{}
	cursor.Next(rawCommit2)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit2)

	require.Equal(t, 2, len(rawCommit2.BranchClocks))          // Belongs to 2 branches
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // Has 1 commit on this branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[1].Clocks)) // Has 1 commit on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[0])
	require.Equal(t, &Clock{Branch: "foo", Clock: 0}, rawCommit2.BranchClocks[1].Clocks[0])

	commit2 := persistCommitToPFSCommit(rawCommit2)
	err = d.FinishCommit(commit2, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

}

func TestStartCommitSpecifyParentAndNoBranch(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"master",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err := gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)
	defer func() {
		require.NoError(t, cursor.Close())
	}()

	rawCommit := &Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0].Clock)

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

	commit2ID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commit2ID,
		commitID,
		"",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err = gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)

	rawCommit2 := &Commit{}
	cursor.Next(rawCommit2)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit2)

	require.Equal(t, 1, len(rawCommit2.BranchClocks))          // Only belongs to one branch
	require.Equal(t, 2, len(rawCommit.BranchClocks[0].Clocks)) // Has 2 commits on this branch
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[0])
	require.Equal(t, &Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[1])

	commit2 := persistCommitToPFSCommit(rawCommit2)
	err = d.FinishCommit(commit2, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

}

func TestStartCommitNoParentOrBranch(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{Name: "foo"},
		commitID,
		"",
		"",
		timestampNow(),
		make([]*pfs.Commit, 0),
		make(map[uint64]bool),
	)
	require.NoError(t, err)

	cursor, err := gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Default(gorethink.Error("value not found")).Run(dbClient)
	defer func() {
		require.NoError(t, cursor.Close())
	}()

	rawCommit := &Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Matches(t, "foo", rawCommit.BranchClocks[0].Clocks[0].Branch)
	require.Equal(t, 0, rawCommit.BranchClocks[0].Clocks[0].Clock)
}
