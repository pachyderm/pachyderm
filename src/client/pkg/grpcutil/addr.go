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
	// allow no specification of a pachd address
	if value == "" {
		return PachdAddress{}, ErrNoPachdAddress
	}

	// parse
	u, err := url.Parse(value)
	if err != nil {
		return PachdAddress{}, fmt.Errorf("could not parse pachd address: %v", err)
	}

	// validate
	if u.Scheme != "" && u.Scheme != "grpc" && u.Scheme != "grpcs" {
		return PachdAddress{}, fmt.Errorf("unrecognized scheme in pachd address: %s", u.Scheme)
	}
	if u.User != nil {
		return PachdAddress{}, errors.New("pachd address should not include login credentials")
	}
	if u.Path != "" {
		return PachdAddress{}, errors.New("pachd address should not include a path")
	}
	if u.RawQuery != "" {
		return PachdAddress{}, errors.New("pachd address should not include a query string")
	}
	if u.Fragment != "" {
		return PachdAddress{}, errors.New("pachd address should not include a fragment")
	}

	// sanitize
	// port always starts after last colon, but net.SplitHostPort returns an
	// error on a hostport without a colon, which this might be
	host := u.Host
	port := ""
	if colonIdx := strings.LastIndexByte(u.Host, ':'); colonIdx >= 0 {
		host = u.Host[:colonIdx]
		port = u.Host[colonIdx+1:]
	}
	var portInt uint64 = DefaultPachdNodePort
	if port != "" {
		portInt, err = strconv.ParseUint(port, 10, 16)
		if err != nil {
			return PachdAddress{}, fmt.Errorf("could not parse port of the pachd address: %v", err)
		}
	}

	return PachdAddress{
		Secured: u.Scheme == "grpcs",
		Host:    host,
		Port:    uint16(portInt),
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
