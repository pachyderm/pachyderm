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
	port int32 = 40651
)

func TestCommitBranchIndexBasicRF(t *testing.T) {
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
		cursor, err = term.GetAllByIndex(CommitBranchIndex.GetName(), key).Run(dbClient)
		require.NoError(t, err)
		returnedCommit := &persist.Commit{}
		require.NoError(t, cursor.One(returnedCommit))
		require.Equal(t, commitID, returnedCommit.ID)
	})
}

func TestCommitBranchIndexHeadOfBranchRF(t *testing.T) {
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
		// master exists, when providing parentID and branch, assume its a new branch
		require.YesError(t, err)

		err = d.StartCommit(
			repo,
			commitID2,
			"",
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
			"",
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
			Index: gorethink.Desc(CommitBranchIndex.GetName()),
		}).Between([]interface{}{repo.Name, "master", 0}, []interface{}{repo.Name, "master", gorethink.MaxVal}, gorethink.BetweenOpts{
			RightBound: "open",
		}).Run(dbClient)
		require.NoError(t, err)
		require.NoError(t, cursor.One(head))
		require.Equal(t, commitID3, head.ID)
	})
}

func TestDiffCommitIndexBasicRF(t *testing.T) {
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

		cursor, err := gorethink.DB(dbName).Table(diffTable).Map(DiffCommitIndex.GetCreateFunction()).Run(dbClient)
		require.NoError(t, err)
		var indexedAsCommit string
		require.NoError(t, cursor.One(&indexedAsCommit))
		require.Equal(t, commitID, indexedAsCommit)
	})
}

func TestDiffPathIndexBasicRF(t *testing.T) {

	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session, client pclient.APIClient) {

		repo := &pfs.Repo{Name: "repo1"}
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

		cursor, err := gorethink.DB(dbName).Table(diffTable).Map(DiffPathIndex.GetCreateFunction()).Run(dbClient)
		require.NoError(t, err)
		fields := []interface{}{}
		// Example return value:
		// []interface {}{"repo1", "foo/bar/fizz/buzz", []interface {}{[]interface {}{"master", 0}}}
		expectedPrefixes := []interface{}{
			"/foo",
			"/foo/bar",
			"/foo/bar/fizz",
			"/foo/bar/fizz/buzz",
		}
		for cursor.Next(&fields) {
			innerFields, ok := fields[0].([]interface{})
			require.Equal(t, true, ok)
			require.Equal(t, repo.Name, innerFields[0].(string))

			path := innerFields[1].(string)
			require.EqualOneOf(t, expectedPrefixes, path)
			for i, expectedPrefix := range expectedPrefixes {
				if expectedPrefix == path {
					expectedPrefixes = append(expectedPrefixes[:i], expectedPrefixes[i+1:]...)
				}
			}

			clock, ok := innerFields[2].([]interface{})
			require.Equal(t, true, ok)
			require.Equal(t, "master", clock[0].(string))
			require.Equal(t, float64(0), clock[1].(float64))
		}
		require.Equal(t, 0, len(expectedPrefixes))
		require.NoError(t, cursor.Err())
	})
}

func TestClockBranchIndexRF(t *testing.T) {

	testSetup(t, func(d drive.Driver, dbName string, dbClient *gorethink.Session, client pclient.APIClient) {

		repo := &pfs.Repo{Name: "repo1"}
		require.NoError(t, d.CreateRepo(repo, timestampNow(), nil, nil))
		err := d.StartCommit(
			repo,
			uuid.NewWithoutDashes(),
			"",
			"master",
			timestampNow(),
			nil,
			nil,
		)
		require.NoError(t, err)
		err = d.StartCommit(
			repo,
			uuid.NewWithoutDashes(),
			"",
			"foo",
			timestampNow(),
			nil,
			nil,
		)
		require.NoError(t, err)
		err = d.StartCommit(
			repo,
			uuid.NewWithoutDashes(),
			"",
			"bar",
			timestampNow(),
			nil,
			nil,
		)
		require.NoError(t, err)

		removeElem := func(elements []interface{}, element interface{}) []interface{} {
			split := 0
			for i, thisElement := range elements {
				if thisElement == element {
					split = i
					break

				}
			}
			return append(elements[:split], elements[(split+1):]...)
		}

		cursor, err := gorethink.DB(dbName).Table(clockTable).Map(ClockBranchIndex.GetCreateFunction()).Run(dbClient)
		require.NoError(t, err)
		fields := []interface{}{}
		branches := []interface{}{"master", "foo", "bar"}
		require.Equal(t, true, cursor.Next(&fields))
		require.Equal(t, "repo1", fields[0].(string))
		thisBranch := fields[1].(string)
		require.EqualOneOf(t, branches, thisBranch)
		branches = removeElem(branches, thisBranch)

		require.Equal(t, true, cursor.Next(&fields))
		require.Equal(t, "repo1", fields[0].(string))
		thisBranch = fields[1].(string)
		require.EqualOneOf(t, branches, thisBranch)
		branches = removeElem(branches, thisBranch)

		require.Equal(t, true, cursor.Next(&fields))
		require.Equal(t, "repo1", fields[0].(string))
		thisBranch = fields[1].(string)
		require.EqualOneOf(t, branches, thisBranch)
		branches = removeElem(branches, thisBranch)

		require.Equal(t, 0, len(branches))

		require.Equal(t, false, cursor.Next(&fields))
		require.NoError(t, cursor.Err())
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
