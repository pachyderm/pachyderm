package persist

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"

	"github.com/dancannon/gorethink"
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

func persistCommitToPFSCommit(rawCommit *persist.Commit) *pfs.Commit {
	return &pfs.Commit{
		Repo: &pfs.Repo{
			Name: rawCommit.Repo,
		},
		ID: rawCommit.ID,
	}
}

func TestStartCommitRF(t *testing.T) {
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

func TestStartCommitRace(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	repo := &pfs.Repo{Name: "foo"}
	require.NoError(t, d.CreateRepo(repo, timestampNow(), nil, nil))

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		repo,
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

		cursor, err = gorethink.DB(RethinkTestDB).Table(commitTable).Get(newCommitID).Run(dbClient)
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
