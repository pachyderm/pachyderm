package grpcutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestParsePachdAddress(t *testing.T) {
	_, err := ParsePachdAddress("")
	require.True(t, errors.Is(err, ErrNoPachdAddress))

	_, err = ParsePachdAddress("grpc://user@pachyderm.com:80")
	require.YesError(t, err)

	_, err = ParsePachdAddress("grpc://user:pass@pachyderm.com:80")
	require.YesError(t, err)

	_, err = ParsePachdAddress("grpc://pachyderm.com:80/")
	require.YesError(t, err)

	_, err = ParsePachdAddress("grpc://pachyderm.com:80/?foo")
	require.YesError(t, err)

	_, err = ParsePachdAddress("grpc://pachyderm.com:80/#foo")
	require.YesError(t, err)

	p, err := ParsePachdAddress("http://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdAddress("https://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdAddress("grpc://pachyderm.com:80")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    80,
	}, p)

	p, err = ParsePachdAddress("grpcs://[::1]:80")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: true,
		Host:    "::1",
		Port:    80,
	}, p)

	p, err = ParsePachdAddress("grpc://pachyderm.com")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}, p)

	p, err = ParsePachdAddress("127.0.0.1")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "127.0.0.1",
		Port:    DefaultPachdNodePort,
	}, p)

	p, err = ParsePachdAddress("127.0.0.1:80")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "127.0.0.1",
		Port:    80,
	}, p)

	p, err = ParsePachdAddress("[::1]")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "::1",
		Port:    DefaultPachdNodePort,
	}, p)

	p, err = ParsePachdAddress("[::1]:80")
	require.NoError(t, err)
	require.Equal(t, &PachdAddress{
		Secured: false,
		Host:    "::1",
		Port:    80,
	}, p)
}

func TestPachdAddressQualified(t *testing.T) {
	p := &PachdAddress{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}

	require.Equal(t, "grpc://pachyderm.com:30650", p.Qualified())
	p = &PachdAddress{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "grpcs://pachyderm.com:30650", p.Qualified())
}

func TestPachdAddressHostname(t *testing.T) {
	p := &PachdAddress{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "pachyderm.com:30650", p.Hostname())

	p = &PachdAddress{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "pachyderm.com:30650", p.Hostname())
}
