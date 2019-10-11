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
}

// ParsePachdAddress parses a string into a pachd address, or returns an error
// if it's invalid
func ParsePachdAddress(value string) (PachdAddress, error) {
	if value == "" {
		return PachdAddress{}, ErrNoPachdAddress
	}

	secured := false

	if strings.Contains(value, "://") {
		// only parse the url if it contains a scheme, as net/url doesn't
		// appropriately handle values without one

		u, err := url.Parse(value)
		if err != nil {
			return PachdAddress{}, fmt.Errorf("could not parse pachd address: %v", err)
		}

		if u.Scheme != "grpc" && u.Scheme != "grpcs" {
			return PachdAddress{}, fmt.Errorf("unrecognized scheme in pachd address: %s", u.Scheme)
		}
		if u.User != nil {
			return PachdAddress{}, errors.New("pachd address should not include login credentials")
		}
		if u.RawQuery != "" {
			return PachdAddress{}, errors.New("pachd address should not include a query string")
		}
		if u.Fragment != "" {
			return PachdAddress{}, errors.New("pachd address should not include a fragment")
		}
		if u.Path != "" {
			return PachdAddress{}, errors.New("pachd address should not include a path")
		}

		value = u.Host
		secured = u.Scheme == "grpcs"
	}

	// port always starts after last colon, but net.SplitHostPort returns an
	// error on a hostport without a colon, which this might be
	colonIdx := strings.LastIndexByte(value, ':')
	host := value
	port := uint16(DefaultPachdNodePort)
	if colonIdx >= 0 {
		maybePort, err := strconv.ParseUint(value[colonIdx+1:], 10, 16)
		if err == nil {
			host = value[:colonIdx]
			port = uint16(maybePort)
		}
	}

	return PachdAddress{
		Secured: secured,
		Host:    host,
		Port:    port,
	}, nil
}

// Qualified returns the "fully qualified" address, including the scheme
func (p PachdAddress) Qualified() string {
	if p.Secured {
		return fmt.Sprintf("grpcs://%s:%d", p.Host, p.Port)
	} else {
		return fmt.Sprintf("grpc://%s:%d", p.Host, p.Port)
	}
}

// Hostname returns the host:port combination of the pachd address, without
// the scheme
func (p PachdAddress) Hostname() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

// IsUnusualPort returns true if the pachd address port is not one of the
// usual values
func (p PachdAddress) IsUnusualPort() bool {
	return p.Port != DefaultPachdNodePort && p.Port != DefaultPachdPort
}

// IsLoopback returns whether the pachd address is referencing the loopback
// hostname
func (p PachdAddress) IsLoopback() bool {
	return p.Host == "0.0.0.0" || p.Host == "127.0.0.1" || p.Host == "[::1]" || p.Host == "localhost"
}
