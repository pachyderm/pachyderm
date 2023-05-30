package grpcutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestParsePachdAddress(t *testing.T) {
	testData := []struct {
		url     string
		want    *PachdAddress
		wantErr bool
	}{
		{
			url:     "",
			wantErr: true,
		},
		{
			url:     "grpc://user@pachyderm.com:80",
			wantErr: true,
		},
		{
			url:     "grpc://user:pass@pachyderm.com:80",
			wantErr: true,
		},
		{
			url:     "grpc://pachyderm.com:80/",
			wantErr: true,
		},
		{
			url:     "grpc://pachyderm.com:80/?foo",
			wantErr: true,
		},
		{
			url:     "grpc://pachyderm.com:80/#foo",
			wantErr: true,
		},
		{
			url:     "invalid://pachyderm.com",
			wantErr: true,
		},
		{
			url: "http://pachyderm.com:80",
			want: &PachdAddress{
				Secured: false,
				Host:    "pachyderm.com",
				Port:    80,
			},
		},
		{
			url: "https://pachyderm.com:80",
			want: &PachdAddress{
				Secured: true,
				Host:    "pachyderm.com",
				Port:    80,
			},
		},
		{
			url: "grpc://pachyderm.com:80",
			want: &PachdAddress{
				Secured: false,
				Host:    "pachyderm.com",
				Port:    80,
			},
		},
		{
			url: "grpcs://pachyderm.com:80",
			want: &PachdAddress{
				Secured: true,
				Host:    "pachyderm.com",
				Port:    80,
			},
		},
		{
			url: "http://pachyderm.com",
			want: &PachdAddress{
				Secured: false,
				Host:    "pachyderm.com",
				Port:    80,
			},
		},
		{
			url: "https://pachyderm.com",
			want: &PachdAddress{
				Secured: true,
				Host:    "pachyderm.com",
				Port:    443,
			},
		},
		{
			url: "pachyderm.com",
			want: &PachdAddress{
				Secured: false,
				Host:    "pachyderm.com",
				Port:    DefaultPachdNodePort,
			},
		},
		{
			url: "grpc://pachyderm.com",
			want: &PachdAddress{
				Secured: false,
				Host:    "pachyderm.com",
				Port:    DefaultPachdNodePort,
			},
		},
		{
			url: "grpcs://pachyderm.com",
			want: &PachdAddress{
				Secured: true,
				Host:    "pachyderm.com",
				Port:    DefaultPachdNodePort,
			},
		},
		{
			url: "grpcs://[::1]:80",
			want: &PachdAddress{
				Secured: true,
				Host:    "::1",
				Port:    80,
			},
		},
		{
			url: "127.0.0.1",
			want: &PachdAddress{
				Secured: false,
				Host:    "127.0.0.1",
				Port:    DefaultPachdNodePort,
			},
		},
		{
			url: "127.0.0.1:80",
			want: &PachdAddress{
				Secured: false,
				Host:    "127.0.0.1",
				Port:    80,
			},
		},
		{
			url: "[::1]",
			want: &PachdAddress{
				Secured: false,
				Host:    "::1",
				Port:    DefaultPachdNodePort,
			},
		},
		{
			url: "[::1]:30650",
			want: &PachdAddress{
				Secured: false,
				Host:    "::1",
				Port:    30650,
			},
		},
		{
			url: "[::1]:80",
			want: &PachdAddress{
				Secured: false,
				Host:    "::1",
				Port:    80,
			},
		},
		{
			url: "unix:///tmp/socket",
			want: &PachdAddress{
				Secured:    false,
				UnixSocket: "unix:///tmp/socket",
			},
		},
	}
	for _, test := range testData {
		t.Run(test.url, func(t *testing.T) {
			got, err := ParsePachdAddress(test.url)
			if test.wantErr {
				require.YesError(t, err, "parsing %q should error", test.url)
				return
			} else {
				require.NoError(t, err, "parsing %q should not error", test.url)
			}
			require.Equal(t, test.want, got)
		})
	}

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

	p = &PachdAddress{
		Secured:    false,
		UnixSocket: "unix:///tmp/socket",
	}
	require.Equal(t, "unix:///tmp/socket", p.Qualified())
}

func TestPachdAddressHostname(t *testing.T) {
	p := &PachdAddress{
		Secured: false,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "dns:///pachyderm.com:30650", p.Target())

	p = &PachdAddress{
		Secured: true,
		Host:    "pachyderm.com",
		Port:    DefaultPachdNodePort,
	}
	require.Equal(t, "dns:///pachyderm.com:30650", p.Target())

	p = &PachdAddress{
		Secured:    false,
		UnixSocket: "unix:///tmp/socket",
	}
	require.Equal(t, "unix:///tmp/socket", p.Target())

}
