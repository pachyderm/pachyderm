package prototest

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

// RunT runs the testFunc in the context of a number of grpc Servers.
func RunT(
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

// RunB runs the benchFunc in the context of a number of grpc Servers.
func RunB(
	b *testing.B,
	numServers int,
	registerFunc func(map[string]*grpc.Server),
	benchFunc func(*testing.B, map[string]*grpc.ClientConn),
) {
	grpcSuite := &grpcSuite{
		numServers:   numServers,
		registerFunc: registerFunc,
	}
	grpcSuite.SetupSuite()
	defer grpcSuite.TearDownSuite()
	benchFunc(b, grpcSuite.clientConns)
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
	ports, err := getPorts(g.numServers)
	requireNoError(g.T(), err)
	for i := 0; i < g.numServers; i++ {
		port := ports[i]
		requireNoError(g.T(), err)
		address := fmt.Sprintf("0.0.0.0:%s", port)
		server := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
		g.servers[address] = server
		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		requireNoError(g.T(), err)
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
		close(g.done)
	}()
	g.clientConns = make(map[string]*grpc.ClientConn)
	for address := range g.servers {
		clientConn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			g.TearDownSuite()
			requireNoError(g.T(), err)
		}
		g.clientConns[address] = clientConn
	}
}

func (g *grpcSuite) KillServer(address string) {
	requireNoError(g.T(), g.clientConns[address].Close())
	delete(g.clientConns, address)
	g.servers[address].Stop()
	delete(g.servers, address)
}

func (g *grpcSuite) TearDownSuite() {
	for address := range g.clientConns {
		g.KillServer(address)
	}
	<-g.done
}

func (g *grpcSuite) TestSuite() {
	g.testFunc(g.T(), g.clientConns)
}

func getPorts(count int) (_ []string, retErr error) {
	ports := make([]string, count)
	for i := 0; i < count; i++ {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := listener.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		address := listener.Addr().String()
		_, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ports[i] = port
	}
	return ports, nil
}

func requireNoError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	if err != nil {
		logMessage(tb, msgAndArgs...)
		tb.Errorf("No error is expected but got %v", err)
	}
}

func logMessage(tb testing.TB, msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 1 {
		tb.Logf(msgAndArgs[0].(string))
	}
	if len(msgAndArgs) > 1 {
		tb.Logf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}
