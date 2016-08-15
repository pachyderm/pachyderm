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

	require.Equal(t, 1, len(rawCommit.FullClock))
	require.Equal(t, &persist.Clock{Branch: "master", Clock: 0}, rawCommit.FullClock[0])

	commit := persistCommitToPFSCommit(rawCommit)
	err = d.FinishCommit(commit, timestampNow(), false, make(map[uint64]bool))
	require.NoError(t, err)

	var wg sync.WaitGroup
	ch := make(chan int, 100)
	errCh := make(chan error, 1)
	startCommitOnHead := func() {
		defer wg.Done()
		var bc *persist.Clock
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

		if len(rawCommit.FullClock) != 1 {
			err = fmt.Errorf("BranchClock %v has incorrect number of Clocks", bc)
			goto SendError
		}
		bc = rawCommit.FullClock[0]
		ch <- int(bc.Clock)

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
