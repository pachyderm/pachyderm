package persist

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"github.com/dancannon/gorethink"
	"go.pedge.io/pb/go/google/protobuf"
)

func testSetup(t *testing.T, testCode func(drive.Driver, string, *gorethink.Session)) {
	dbName := "pachyderm_test_" + uuid.NewWithoutDashes()[0:12]
	if err := InitDB(RethinkAddress, dbName); err != nil {
		require.NoError(t, err)
		return
	}
	dbClient, err := gorethink.Connect(gorethink.ConnectOpts{
		Address: RethinkAddress,
		Timeout: connectTimeoutSeconds * time.Second,
	})
	require.NoError(t, err)
	d, err := NewDriver("localhost:1523", RethinkAddress, dbName)
	require.NoError(t, err)

	testCode(d, dbName, dbClient)

	if err := RemoveDB(RethinkAddress, dbName); err != nil {
		require.NoError(t, err)
		return
	}
}

/*
	CommitBranchIndex

	Given a repo and a clock, returns a commit

	Is used in several places to:

	- find a parent commit given the parent's id in the form of an alias (e.g. "master/0")
	- getHeadOfBranch() -- by doing a range query of the form "branchName/0" to "branchName/max" and returning the last result (in this case the head)
	- getIDOfParentcommit() -- by decrementing this commit's clock value, and searching for that new clock
	- getCommitByAmbmiguousID() -- if the commit ID is in the form of an alias, find the commit using the index

*/
func TestCommitBranchIndexBasic(t *testing.T) {
	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session) {

		repo := &pfs.Repo{Name: "foo"}
		require.NoError(t, d.CreateRepo(repo, timestampNow(), nil, nil))
		commitID := uuid.NewWithoutDashes()
		err := d.StartCommit(
			repo,
			commitID,
			"",
			"master",
			timestampNow(),
			nil,
			nil,
		)
		require.NoError(t, err)

		term := gorethink.DB(dbName).Table(commitTable)
		cursor, err := term.GetAll(commitID).Run(dbClient)
		require.NoError(t, err)
		c := &persist.Commit{}
		err = cursor.One(c)
		require.NoError(t, err)
		require.Equal(t, commitID, c.ID)

		clock := &persist.Clock{Branch: "master", Clock: 0}
		clockID := getClockID(repo.Name, clock).ID
		cursor, err = gorethink.DB(dbName).Table(clockTable).GetAll(clockID).Run(dbClient)
		require.NoError(t, err)
		returnedClock := &persist.Clock{}
		require.NoError(t, cursor.One(returnedClock))
		require.Equal(t, "master", returnedClock.Branch)
		require.Equal(t, uint64(0), returnedClock.Clock)

		key := []interface{}{repo.Name, clock.Branch, clock.Clock}
		cursor, err = term.GetAllByIndex(commitBranchIndex, key).Run(dbClient)
		require.NoError(t, err)
		returnedCommit := &persist.Commit{}
		require.NoError(t, cursor.One(returnedCommit))
		require.Equal(t, commitID, returnedCommit.ID)
	})
}

/*
	diffPathIndex

	used nowhere?
*/

/* diffCommitIndex

Indexed on commitID field in diff row

Used to:

- in FinishCommit() to gather all of the diffs for this commit

*/

/* commitModifiedPathsIndex

Indexed on: repo / path / branchclock

Used to:

- in GetFile() to gather all commits within a range that modified a given path

*/

/* commitDeletedPathsIndex

Indexed on :

Used to:

- in GetFile() given a range of commits, find the commits that deleted a path, return the leftmost one

*/

/* clockBranchIndex

Indexed on:

- repo && branchIndex

Used to:

- in ListBranch() to query the clocks table and return the branches

*/

func timestampNow() *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{Seconds: time.Now().Unix()}
}
