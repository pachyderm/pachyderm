package persist

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"

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

func persistCommitToPFSCommit(rawCommit *persist.Commit) *pfs.Commit {
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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0])

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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0])

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
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 1}, rawCommit.BranchClocks[0].Clocks[0])

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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0].Clock)

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

	rawCommit2 := &persist.Commit{}
	cursor.Next(rawCommit2)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit2)

	require.Equal(t, 1, len(rawCommit2.BranchClocks))          // Only belongs to one branch
	require.Equal(t, 2, len(rawCommit.BranchClocks[0].Clocks)) // Has 2 commits on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[0])
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 1}, rawCommit2.BranchClocks[0].Clocks[1])

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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0].Clock)

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

	rawCommit2 := &persist.Commit{}
	cursor.Next(rawCommit2)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit2)

	require.Equal(t, 2, len(rawCommit2.BranchClocks))          // Belongs to 2 branches
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // Has 1 commit on this branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[1].Clocks)) // Has 1 commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[0])
	require.Equal(t, &persist.Clock{Branch: "foo", Clock: 0}, rawCommit2.BranchClocks[1].Clocks[0])

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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0].Clock)

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

	rawCommit2 := &persist.Commit{}
	cursor.Next(rawCommit2)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit2)

	require.Equal(t, 1, len(rawCommit2.BranchClocks))          // Only belongs to one branch
	require.Equal(t, 2, len(rawCommit.BranchClocks[0].Clocks)) // Has 2 commits on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[0])
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit2.BranchClocks[0].Clocks[1])

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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Matches(t, "foo", rawCommit.BranchClocks[0].Clocks[0].Branch)
	require.Equal(t, 0, rawCommit.BranchClocks[0].Clocks[0].Clock)
}

func TestStartCommitRace(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)

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

	rawCommit := &persist.Commit{}
	cursor.Next(rawCommit)
	require.NoError(t, cursor.Err())

	fmt.Printf("Commit info: %v\n", rawCommit)

	require.Equal(t, 1, len(rawCommit.BranchClocks))           // Only belongs to one branch
	require.Equal(t, 1, len(rawCommit.BranchClocks[0].Clocks)) // First commit on this branch
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.BranchClocks[0].Clocks[0])

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

	var wg sync.WaitGroup
	ch := make(chan int, 100)
	errCh := make(chan error, 1)
	startCommitOnHead := func() {
		defer wg.Done()
		var bc *persist.BranchClock
		newCommitID := uuid.NewWithoutDashes()
		err := d.StartCommit(
			&pfs.Repo{Name: "foo"},
			newCommitID,
			"",
			"master",
			timestampNow(),
			make([]*pfs.Commit, 0),
			make(map[uint64]bool),
		)
		if err != nil {
			goto SendError
		}

		cursor, err = gorethink.DB(RethinkTestDB).Table(commitTable).Get(commitID).Run(dbClient)
		if err != nil {
			goto SendError
		}

		if err = cursor.One(rawCommit); err != nil {
			goto SendError
		}

		if len(rawCommit.BranchClocks) != 1 {
			err = fmt.Errorf("RawCommit %v has incorrect number of BranchClocks", rawCommit)
			goto SendError
		}

		bc = rawCommit.BranchClocks[0]
		if len(bc.Clocks) != 1 {
			err = fmt.Errorf("BranchClock %v has incorrect number of Clocks", bc)
			goto SendError
		}
		ch <- int(bc.Clocks[len(bc.Clocks)-1].Clock)

	SendError:
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go startCommitOnHead()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		require.NoError(t, err)
	default:
	}

	results := make(map[int]bool)
	for j := 0; j < 100; j++ {
		value := <-ch
		results[value] = true
	}

	// There should be a bunch of new commits and no collisions
	require.Equal(t, 100, len(results))
}
