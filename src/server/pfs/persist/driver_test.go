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

func TestStartCommit(t *testing.T) {
	d, err := NewDriver("localhost:1523", RethinkAddress, RethinkTestDB)
	require.NoError(t, err)
	fmt.Printf("got a driver")

	dbClient, err := dbConnect(RethinkAddress)
	require.NoError(t, err)

	commitID := uuid.NewWithoutDashes()
	err = d.StartCommit(
		&pfs.Repo{},
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

}
