package grpcutil

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	// DefaultPachdNodePort is the pachd kubernetes service's default
	// NodePort.Port setting
	DefaultPachdNodePort = 30650
)

var (
	// ErrNoPachdAddress is returned by ParsePachdAddress when the input is an
	// empty string
	ErrNoPachdAddress = errors.New("no pachd address specified")
	// DefaultPachdAddress is the default PachdAddress that should be used
	// if none is otherwise specified. It's a loopback that should rely on
	// port forwarding.
	DefaultPachdAddress = PachdAddress{
		Secured: false,
		Host:    "0.0.0.0",
		Port:    DefaultPachdNodePort,
	}
)

// PachdAddress represents a parsed pachd address value
type PachdAddress struct {
	// Secured specifies whether grpcs should be used
	Secured bool
	// Host specifies the pachd address host without the port
	Host string
	// Port specifies the pachd port
	Port uint16
	// UnixSocket is set if the pachd address refers to a unix socket
	UnixSocket string
}

// ParsePachdAddress parses a string into a pachd address, or returns an error
// if it's invalid
func ParsePachdAddress(value string) (*PachdAddress, error) {
	if value == "" {
		return nil, ErrNoPachdAddress
	}

	if !strings.Contains(value, "://") {
		// append a default scheme if one does not exist, as `url.Parse`
		// doesn't appropriately handle values without one
		value = "grpc://" + value
	}

	u, err := url.Parse(value)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse pachd address")
	}

	switch u.Scheme {
	case "grpc", "grpcs", "http", "https":
	case "unix":
		return &PachdAddress{UnixSocket: value}, nil
	default:
		return nil, errors.Errorf("unrecognized scheme in pachd address: %s", u.Scheme)
	}

	switch {
	case u.Path != "":
		return nil, errors.New("pachd address should not include a path")
	case u.User != nil:
		return nil, errors.New("pachd address should not include login credentials")
	case u.RawQuery != "":
		return nil, errors.New("pachd address should not include a query string")
	case u.Fragment != "":
		return nil, errors.New("pachd address should not include a fragment")
	}

	port := uint16(DefaultPachdNodePort)
	if strport := u.Port(); strport != "" {
		maybePort, err := strconv.ParseUint(strport, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse port in address")
		}
		port = uint16(maybePort)
	}

	return &PachdAddress{
		Secured: u.Scheme == "grpcs" || u.Scheme == "https",
		Host:    u.Hostname(),
		Port:    port,
	}, nil
}

// Qualified returns the "fully qualified" address, including the scheme
func (p *PachdAddress) Qualified() string {
	if p.Secured {
		return fmt.Sprintf("grpcs://%s:%d", p.Host, p.Port)
	}
	if p.UnixSocket != "" {
		return p.UnixSocket
	}
	return fmt.Sprintf("grpc://%s:%d", p.Host, p.Port)
}

// Target returns a string suitable for calling grpc.Dial.
// This may be a host:port pair for TCP connections, or a unix socket address.
func (p *PachdAddress) Target() string {
	if p.UnixSocket != "" {
		return p.UnixSocket
	}
	return fmt.Sprintf("dns:///%s:%d", p.Host, p.Port)
}
