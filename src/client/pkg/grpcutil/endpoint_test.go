package grpcutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestParsePachdEndpointTCP(t *testing.T) {
	_, err := ParsePachdEndpoint("")
	require.Equal(t, err, ErrNoPachdEndpoint)

	p, err := ParsePachdEndpoint("127.0.0.1")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://127.0.0.1:30650")
	require.Equal(t, p.Address(), "127.0.0.1:30650")
	require.False(t, p.Secured())

	p, err = ParsePachdEndpoint("127.0.0.1:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://127.0.0.1:80")
	require.Equal(t, p.Address(), "127.0.0.1:80")
	require.False(t, p.Secured())

	p, err = ParsePachdEndpoint("[::1]")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://::1:30650")
	require.Equal(t, p.Address(), "::1:30650")
	require.False(t, p.Secured())

	p, err = ParsePachdEndpoint("[::1]:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://::1:80")
	require.Equal(t, p.Address(), "::1:80")
	require.False(t, p.Secured())

	_, err = ParsePachdEndpoint("grpc://user@pachyderm.com:80")
	require.YesError(t, err)

	_, err = ParsePachdEndpoint("grpc://user:pass@pachyderm.com:80")
	require.YesError(t, err)

	_, err = ParsePachdEndpoint("grpc://pachyderm.com:80/")
	require.YesError(t, err)

	_, err = ParsePachdEndpoint("grpc://pachyderm.com:80/?foo")
	require.YesError(t, err)

	_, err = ParsePachdEndpoint("grpc://pachyderm.com:80/#foo")
	require.YesError(t, err)

	p, err = ParsePachdEndpoint("grpc://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://pachyderm.com:80")
	require.Equal(t, p.Address(), "pachyderm.com:80")
	require.False(t, p.Secured())

	p, err = ParsePachdEndpoint("grpc://pachyderm.com")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://pachyderm.com:30650")
	require.Equal(t, p.Address(), "pachyderm.com:30650")
	require.False(t, p.Secured())

	p, err = ParsePachdEndpoint("http://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://pachyderm.com:80")
	require.Equal(t, p.Address(), "pachyderm.com:80")
	require.False(t, p.Secured())

	p, err = ParsePachdEndpoint("tcp://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpc://pachyderm.com:80")
	require.Equal(t, p.Address(), "pachyderm.com:80")
	require.False(t, p.Secured())
}

func TestParsePachdEndpointTCPS(t *testing.T) {
	p, err := ParsePachdEndpoint("https://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpcs://pachyderm.com:80")
	require.Equal(t, p.Address(), "pachyderm.com:80")
	require.True(t, p.Secured())

	p, err = ParsePachdEndpoint("tcps://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpcs://pachyderm.com:80")
	require.Equal(t, p.Address(), "pachyderm.com:80")
	require.True(t, p.Secured())

	p, err = ParsePachdEndpoint("tls://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpcs://pachyderm.com:80")
	require.Equal(t, p.Address(), "pachyderm.com:80")
	require.True(t, p.Secured())

	p, err = ParsePachdEndpoint("grpcs://[::1]:80")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "grpcs://::1:80")
	require.Equal(t, p.Address(), "::1:80")
	require.True(t, p.Secured())
}

func TestParsePachdEndpointUDS(t *testing.T) {
	p, err := ParsePachdEndpoint("uds:///tmp/foo")
	require.NoError(t, err)
	require.Equal(t, p.URL(), "uds:///tmp/foo")
	require.Equal(t, p.Address(), "/tmp/foo")
	require.False(t, p.Secured())
}

func TestTCPPachdEndpointIsUnusualPort(t *testing.T) {
	p := &TCPPachdEndpoint{
		Host: "pachyderm.com",
		Port: DefaultPachdNodePort,
	}
	require.False(t, p.IsUnusualPort())

	p = &TCPPachdEndpoint{
		Host: "pachyderm.com",
		Port: DefaultPachdPort,
	}
	require.False(t, p.IsUnusualPort())

	p = &TCPPachdEndpoint{
		Host: "pachyderm.com",
		Port: 80,
	}
	require.True(t, p.IsUnusualPort())
}

func TestTCPPachdEndpointIsLoopback(t *testing.T) {
	p := &TCPPachdEndpoint{
		Host: "localhost",
		Port: DefaultPachdPort,
	}
	require.True(t, p.IsLoopback())

	p = &TCPPachdEndpoint{
		Host: "pachyderm.com",
		Port: DefaultPachdPort,
	}
	require.False(t, p.IsLoopback())
}
