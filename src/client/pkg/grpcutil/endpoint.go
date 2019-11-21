package grpcutil

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// DefaultPachdNodePort is the pachd kubernetes service's default
	// NodePort.Port setting
	DefaultPachdNodePort = 30650

	// DefaultPachdPort is the pachd kubernetes service's default
	// Port (often used with Pachyderm ELBs)
	DefaultPachdPort = 650
)

var (
	// ErrNoPachdEndpoint is returned by ParsePachdEndpoint when the input is an
	// empty string
	ErrNoPachdEndpoint = errors.New("no pachd address specified")
	// DefaultPachdEndpoint is the default PachdEndpoint that should be used
	// if none is otherwise specified. It's a TCP loopback that should rely on
	// port forwarding.
	DefaultPachdEndpoint = TCPPachdEndpoint{
		secured: false,
		Host:    "0.0.0.0",
		Port:    DefaultPachdNodePort,
	}
)

// PachdEndpoint is a structured reference to a pachd instance. Each supported
// underlying protocol (e.g. TCP, UDS) implements this interface.
type PachdEndpoint interface {
	// URL returns a URL version of the pachd endpoint
	URL() string

	// Address returns the pachd address without the scheme
	Address() string

	// Secured returns whether the endpoint is using TLS
	Secured() bool
}

// ParsePachdEndpoint parses a string into a pachd endpoint, or returns an
// error if it's invalid
func ParsePachdEndpoint(value string) (PachdEndpoint, error) {
	if value == "" {
		return nil, ErrNoPachdEndpoint
	}

	if !strings.Contains(value, "://") {
		// append a default scheme if one does not exist, as `url.Parse`
		// doesn't appropriately handle values without one
		value = "grpc://" + value
	}

	u, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("could not parse pachd endpoint: %v", err)
	}

	switch u.Scheme {
	case "grpc", "grpcs", "http", "https", "tcp", "tcps", "tls", "uds":
	default:
		return nil, fmt.Errorf("unrecognized scheme in pachd endpoint: %s", u.Scheme)
	}

	switch {
	case u.User != nil:
		return nil, errors.New("pachd endpoint should not include login credentials")
	case u.RawQuery != "":
		return nil, errors.New("pachd endpoint should not include a query string")
	case u.Fragment != "":
		return nil, errors.New("pachd endpoint should not include a fragment")
	}

	if u.Scheme == "uds" {
		if u.Path == "" {
			return nil, errors.New("UDS pachd endpoint should include a path")
		}
		return &UDSPachdEndpoint{
			Path: u.Path,
		}, nil
	}

	if u.Path != "" {
		return nil, errors.New("pachd endpoint should not include a path")
	}

	port := uint16(DefaultPachdNodePort)
	if strport := u.Port(); strport != "" {
		maybePort, err := strconv.ParseUint(strport, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("could not parse port in pachd endpoint: %v", err)
		}
		port = uint16(maybePort)
	}

	return &TCPPachdEndpoint{
		secured: u.Scheme == "grpcs" || u.Scheme == "https" || u.Scheme == "tcps" || u.Scheme == "tls",
		Host:    u.Hostname(),
		Port:    port,
	}, nil
}

// TCPPachdEndpoint represents a TCP-based pachd endpoint
type TCPPachdEndpoint struct {
	// secured specifies whether grpcs should be used
	secured bool
	// Host specifies the pachd endpoint host without the port
	Host string
	// Port specifies the pachd port
	Port uint16
}

// URL returns a URL version of the pachd endpoint
func (p *TCPPachdEndpoint) URL() string {
	if p.secured {
		return fmt.Sprintf("grpcs://%s:%d", p.Host, p.Port)
	}
	return fmt.Sprintf("grpc://%s:%d", p.Host, p.Port)
}

// Address returns the pachd address without the scheme
func (p *TCPPachdEndpoint) Address() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

// Secured returns whether the endpoint is using TLS
func (p *TCPPachdEndpoint) Secured() bool {
	return p.secured
}

// IsUnusualPort returns true if the port is not one of the usual values
func (p *TCPPachdEndpoint) IsUnusualPort() bool {
	return p.Port != DefaultPachdNodePort && p.Port != DefaultPachdPort
}

// IsLoopback returns whether the host is a loopback
func (p *TCPPachdEndpoint) IsLoopback() bool {
	return p.Host == "0.0.0.0" || p.Host == "127.0.0.1" || p.Host == "[::1]" || p.Host == "localhost"
}

// UDSPachdEndpoint represents a UDS-based pachd endpoint
type UDSPachdEndpoint struct {
	// Path represents the path the UDS socket
	Path string
}

// URL returns a URL version of the pachd endpoint
func (p *UDSPachdEndpoint) URL() string {
	return fmt.Sprintf("uds://%s", p.Path)
}

// Address returns the pachd address without the scheme
func (p *UDSPachdEndpoint) Address() string {
	return p.Path
}

// Secured returns whether the endpoint is using TLS
func (p *UDSPachdEndpoint) Secured() bool {
	return false
}
