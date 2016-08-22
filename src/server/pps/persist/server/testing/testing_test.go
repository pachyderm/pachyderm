package testing

import (
	"fmt"
	"sync/atomic"
	"testing"

	"go.pedge.io/proto/server"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
	"github.com/pachyderm/pachyderm/src/server/pps/persist/server"
)

var (
	port int32 = 38000
)

func RunTestWithPersistClient(t *testing.T, testFunc func(t *testing.T, persistClient persist.APIClient)) {
	if testing.Short() {
		t.Skip("Skipping test because of short mode.")
	}

	apiServer, err := NewTestRethinkAPIServer()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, apiServer.Close())
	}()
	localPort := atomic.AddInt32(&port, 1)
	address := fmt.Sprintf("localhost:%d", localPort)
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				persist.RegisterAPIServer(s, apiServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: version.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	testFunc(t, persist.NewAPIClient(clientConn))
}

func NewTestRethinkAPIServer() (server.APIServer, error) {
	address := "0.0.0.0:32081"
	databaseName := uuid.NewWithoutDashes()
	if err := server.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return server.NewRethinkAPIServer(address, databaseName)
}
