package store

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dancannon/gorethink"
	"github.com/pachyderm/pachyderm/src/common"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	runTest(t, testBasic)
}

func testBasic(session *gorethink.Session, database string) error {
	return nil
}

func runTest(t *testing.T, testFunc func(*gorethink.Session) error) {
	database := strings.Replace(common.NewUUID(), "-", "", -1)

	session, err := getRethinkSession("")
	require.NoError(t, err)
	err = gorethink.DBCreate(database).Exec(session)
	tableErr := initTables(session, database)
	_ = session.Close()
	require.NoError(t, err)
	require.NoError(t, tableErr)

	session, err = getRethinkSession(database)
	testErr := testFunc(session, database)
	_ = session.Close()

	session, err = getRethinkSession("")
	if err == nil {
		_ = gorethink.DBDrop(database).Exec(session)
	}
	_ = session.Close()

	require.NoError(t, testErr)
}

func getRethinkSession(database string) (*gorethink.Session, error) {
	address, err := getRethinkAddress()
	if err != nil {
		return nil, err
	}
	return gorethink.Connect(
		gorethink.ConnectOpts{
			Address:  address,
			Database: database,
		},
	)
}

func getRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
