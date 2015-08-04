package store

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/dancannon/gorethink"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	runTest(t, testBasic)
}

func testBasic(session *gorethink.Session, term gorethink.Term) error {
	return nil
}

func runTest(t *testing.T, testFunc func(*gorethink.Session, gorethink.Term) error) {
	session, err := getRethinkSession()
	require.NoError(t, err)
	err := testFunc(session, gorethink.Database("test"))
	_ = session.Close()
	require.NoError(t, err)
}

func getRethinkRession() (*gorethink.Session, error) {
	address, err := getRethinkAddress()
	if err != nil {
		return nil, err
	}
	return gorethink.Connect(
		gorethink.ConnectOpts{
			Address:  address,
			Databese: "test",
		},
	)
}

func getRethinkAddress() (string, error) {
	rethinkPort := os.Getenv("RETHINK_PORT")
	if rethinkPort == "" {
		return "", errors.New("RETHINK_PORT not set")
	}
	return strings.Replace(rethinkPort, "tcp://", "", -1), nil
}
