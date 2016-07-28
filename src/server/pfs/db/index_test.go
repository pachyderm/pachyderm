package persist

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"

	"github.com/dancannon/gorethink"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/server"
	"google.golang.org/grpc"
)

var (
	port int32 = 30651
)

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
	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session, client pclient.APIClient) {

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

func TestCommitBranchIndexHeadOfBranch(t *testing.T) {
	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session, client pclient.APIClient) {

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
		commit := &pfs.Commit{Repo: repo, ID: commitID}
		require.NoError(t, d.FinishCommit(commit, timestampNow(), false, nil))

		commitID2 := uuid.NewWithoutDashes()
		err = d.StartCommit(
			repo,
			commitID2,
			commitID,
			"master",
			timestampNow(),
			nil,
			nil,
		)
		require.NoError(t, err)
		commit2 := &pfs.Commit{Repo: repo, ID: commitID2}
		require.NoError(t, d.FinishCommit(commit2, timestampNow(), false, nil))

		commitID3 := uuid.NewWithoutDashes()
		err = d.StartCommit(
			repo,
			commitID3,
			commitID2,
			"master",
			timestampNow(),
			nil,
			nil,
		)
		require.NoError(t, err)
		commit3 := &pfs.Commit{Repo: repo, ID: commitID3}
		require.NoError(t, d.FinishCommit(commit3, timestampNow(), false, nil))

		// Now that the commits are chained together,
		// Grab the head of master branch using the index
		head := &persist.Commit{}
		term := gorethink.DB(dbName).Table(commitTable)
		cursor, err := term.OrderBy(gorethink.OrderByOpts{
			Index: gorethink.Desc(commitBranchIndex),
		}).Between([]interface{}{repo.Name, "master", 0}, []interface{}{repo.Name, "master", gorethink.MaxVal}, gorethink.BetweenOpts{
			RightBound: "open",
		}).Run(dbClient)
		require.NoError(t, err)
		require.NoError(t, cursor.One(head))
		require.Equal(t, commitID3, head.ID)
	})
}

/*
	diffPathIndex

	used nowhere?
	used in ListFile() to list diffs by path
*/

/* diffCommitIndex

Indexed on commitID field in diff row

Used to:

- in FinishCommit() to gather all of the diffs for this commit

*/

func TestDiffCommitIndexBasic(t *testing.T) {
	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session, client pclient.APIClient) {

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
		file := &pfs.File{
			Commit: &pfs.Commit{
				Repo: repo,
				ID:   commitID,
			},
			Path: "file",
		}
		d.PutFile(file, "", pfs.Delimiter_LINE, 0, strings.NewReader("foo\n"))

		commit := &pfs.Commit{Repo: repo, ID: commitID}
		require.NoError(t, d.FinishCommit(commit, timestampNow(), false, nil))

		cursor, err := gorethink.DB(dbName).Table(diffTable).GetAllByIndex(diffCommitIndex, commitID).Run(dbClient)
		require.NoError(t, err)
		diff := &persist.Diff{}
		require.NoError(t, cursor.One(diff))
		fmt.Printf("got first diff: %v\n", diff)
		require.Equal(t, "file", diff.Path)
		require.Equal(t, 1, len(diff.BlockRefs))

		block := diff.BlockRefs[0]
		fmt.Printf("block: %v\n", block)
		blockSize := block.Upper - block.Lower

		// Was trying to check on a per block level ...
		// But even GetFile() doesn't seem to return the correct results
		// reader, err := d.GetFile(file, nil, 0,
		//	int64(blockSize), &pfs.Commit{Repo: repo, ID: commitID}, 0, false, "")

		reader, err := client.GetBlock(block.Hash, uint64(0), uint64(blockSize))
		require.NoError(t, err)
		var data []byte
		size, err := reader.Read(data)
		fmt.Printf("data=%v, err=%v\n", string(data), err)
		fmt.Printf("size=%v\n", size)
		require.NoError(t, err)
		require.Equal(t, "foo\n", string(data))

	})
}

/* clockBranchIndex

Indexed on:

- repo && branchIndex

Used to:

- in ListBranch() to query the clocks table and return the branches

*/

func TestDiffPathIndexBasic(t *testing.T) {

	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session, client pclient.APIClient) {

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
		file := &pfs.File{
			Commit: &pfs.Commit{
				Repo: repo,
				ID:   commitID,
			},
			Path: "foo/bar/fizz/buzz",
		}
		d.PutFile(file, "", pfs.Delimiter_LINE, 0, strings.NewReader("aaa\n"))

		commit := &pfs.Commit{Repo: repo, ID: commitID}
		require.NoError(t, d.FinishCommit(commit, timestampNow(), false, nil))

		branchClock := &persist.BranchClock{
			Clocks: []*persist.Clock{
				{
					Branch: "master",
					Clock:  0,
				},
			},
		}
		key := []interface{}{repo.Name, false, "foo/bar/fizz/buzz", branchClock.ToArray()}
		cursor, err := gorethink.DB(dbName).Table(diffTable).GetAllByIndex(diffPathIndex.Name, key).Map(diffPathIndex.CreateFunction).Run(dbClient)

		cursor, err = gorethink.DB(dbName).Table(diffTable).Map(diffPathIndex.CreateFunction).Run(dbClient)
		require.NoError(t, err)
		diff := &persist.Diff{}
		require.NoError(t, cursor.One(diff))
		fmt.Printf("got first diff: %v\n", diff)
		require.Equal(t, "file", diff.Path)
		require.Equal(t, 1, len(diff.BlockRefs))
	})
}

func timestampNow() *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{Seconds: time.Now().Unix()}
}

func testSetup(t *testing.T, testCode func(drive.Driver, string, *gorethink.Session, pclient.APIClient)) {
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
	_client, d := startBlockServerAndGetClient(t, dbName)

	testCode(d, dbName, dbClient, _client)

	if err := RemoveDB(RethinkAddress, dbName); err != nil {
		require.NoError(t, err)
		return
	}
}

func runServers(t *testing.T, port int32, blockAPIServer pfsclient.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: version.Version},
			protoserver.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func startBlockServerAndGetClient(t *testing.T, dbName string) (pclient.APIClient, drive.Driver) {

	root := uniqueString("/tmp/pach_test/run")
	var ports []int32
	ports = append(ports, atomic.AddInt32(&port, 1))
	var addresses []string
	for _, port := range ports {
		addresses = append(addresses, fmt.Sprintf("localhost:%d", port))
	}
	var driver drive.Driver
	for i, port := range ports {
		address := addresses[i]
		_driver, err := NewDriver(address, RethinkAddress, dbName)
		driver = _driver
		require.NoError(t, err)
		blockAPIServer, err := pfsserver.NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		runServers(t, port, blockAPIServer)
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pclient.APIClient{BlockAPIClient: pfsclient.NewBlockAPIClient(clientConn)}, driver
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
