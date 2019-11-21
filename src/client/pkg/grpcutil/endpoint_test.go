package grpcutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestParsePachdEndpointTCP(t *testing.T) {
	_, err := ParsePachdEndpoint("")
	require.Equal(t, err, ErrNoPachdEndpoint)

	p, err = ParsePachdEndpoint("127.0.0.1")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "127.0.0.1",
		Port:    DefaultPachdNodePort,
	}, p)

	p, err = ParsePachdEndpoint("127.0.0.1:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "127.0.0.1",
		Port:    80,
	}, p)

	p, err = ParsePachdEndpoint("[::1]")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "::1",
		Port:    DefaultPachdNodePort,
	}, p)

	p, err = ParsePachdEndpoint("[::1]:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "::1",
		Port:    80,
	}, p)

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
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdEndpoint("grpc://pachyderm.com")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}, p)

	p, err := ParsePachdEndpoint("http://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err := ParsePachdEndpoint("tcp://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)
}

func TestParsePachdEndpointTCPS(t *testing.T) {
	p, err := ParsePachdEndpoint("https://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdEndpoint("tcps://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdEndpoint("tls://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdEndpoint("grpcs://[::1]:80")
	require.NoError(t, err)
	require.Equal(t, &TCPPachdEndpoint{
		Secured: true,
		Host:    "::1",
		Port:    80,
	}, p)
}

func TestParsePachdEndpointUDS(t *testing.T) {
	p, err := ParsePachdEndpoint("uds:///tmp/foo")
	require.NoError(t, err)
	require.Equal(t, &UDSPachdEndpoint{
		Path: "/tmp/foo",
	}, p)
}

func TestPachdEndpointURL(t *testing.T) {
	p := &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "grpc://pachyderm.com:30650", p.URL())

	p = &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "grpcs://pachyderm.com:30650", p.URL())

	p = &UDSPachdEndpoint{
		Path: "/tmp/foo",
	}
	require.Equal(t, "uds:///tmp/foo", p.URL())
}

func TestPachdEndpointAddress(t *testing.T) {
	p := &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "pachyderm.com:30650", p.Address())

	p = &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "pachyderm.com:30650", p.Address())

	p = &UDSPachdEndpoint{
		Path: "/tmp/foo",
	}
	require.Equal(t, "/tmp/foo", p.Address())
}

func TestTCPPachdEndpointIsUnusualPort(t *testing.T) {
	p := &TCPPachdEndpoint{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.False(t, p.IsUnusualPort())

	p = &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdPort,
	}
	require.False(t, p.IsUnusualPort())

	p = &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    80,
	}
	require.True(t, p.IsUnusualPort())
}

func TestTCPPachdEndpointIsLoopback(t *testing.T) {
	p := &TCPPachdEndpoint{
		Secured: true,
		Host:    "localhost",
		Port:    DefaultPachdPort,
	}
	require.True(t, p.IsLoopback())

	p = &TCPPachdEndpoint{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdPort,
	}
	require.False(t, p.IsLoopback())
}
