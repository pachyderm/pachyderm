package grpctest

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/facebookgo/freeport"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

func Run(
	t *testing.T,
	numServers int,
	registerFunc func(map[string]*grpc.Server),
	testFunc func(*testing.T, map[string]*grpc.ClientConn),
) {
	grpcSuite := &grpcSuite{
		numServers:   numServers,
		registerFunc: registerFunc,
		testFunc:     testFunc,
	}
	suite.Run(t, grpcSuite)
}

type grpcSuite struct {
	suite.Suite
	numServers   int
	registerFunc func(map[string]*grpc.Server)
	testFunc     func(*testing.T, map[string]*grpc.ClientConn)
	clientConns  map[string]*grpc.ClientConn
	servers      map[string]*grpc.Server
	errC         chan error
	done         chan bool
}

func (g *grpcSuite) SetupSuite() {
	g.servers = make(map[string]*grpc.Server)
	listeners := make(map[string]net.Listener)
	for i := 0; i < g.numServers; i++ {
		port, err := freeport.Get()
		require.NoError(g.T(), err)
		address := fmt.Sprintf("0.0.0.0:%d", port)
		g.servers[address] = grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		require.NoError(g.T(), err)
		listeners[address] = listener
	}
	g.registerFunc(g.servers)
	g.errC = make(chan error, g.numServers)
	for address, server := range g.servers {
		address := address
		server := server
		go func() {
			g.errC <- server.Serve(listeners[address])
		}()
	}
	g.done = make(chan bool, 1)
	go func() {
		for j := 0; j < g.numServers; j++ {
			<-g.errC
		}
		g.done <- true
	}()
	g.clientConns = make(map[string]*grpc.ClientConn)
	for address := range g.servers {
		clientConn, err := grpc.Dial(address)
		if err != nil {
			g.TearDownSuite()
			require.NoError(g.T(), err)
		}
		g.clientConns[address] = clientConn
	}
}

func (g *grpcSuite) TearDownSuite() {
	for _, server := range g.servers {
		server.Stop()
	}
	<-g.done
	for _, clientConn := range g.clientConns {
		_ = clientConn.Close()
	}
}

func (g *grpcSuite) TestSuite() {
	g.testFunc(g.T(), g.clientConns)
}
