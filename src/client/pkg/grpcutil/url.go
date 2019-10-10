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
	ErrNoAddress        = errors.New("no pachd address specified")
	DefaultPachdAddress = PachdAddress{
		Secured: false,
		Host:    "0.0.0.0",
		Port:    DefaultPachdNodePort,
	}
)

type PachdAddress struct {
	Secured bool
	Host    string
	Port    uint16
}

func ParsePachdAddress(value string) (PachdAddress, error) {
	// allow no specification of a pachd address
	if value == "" {
		return PachdAddress{}, ErrNoAddress
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

func (p PachdAddress) Qualified() string {
	if p.Secured {
		return fmt.Sprintf("grpcs://%s:%d", p.Host, p.Port)
	} else {
		return fmt.Sprintf("grpc://%s:%d", p.Host, p.Port)
	}
}

func (p PachdAddress) Hostname() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

func (p PachdAddress) IsUnusualPort() bool {
	return p.Port != DefaultPachdNodePort && p.Port != DefaultPachdPort
}

func (p PachdAddress) IsLoopback() bool {
	return p.Host == "0.0.0.0" || p.Host == "127.0.0.1" || p.Host == "[::1]" || p.Host == "localhost"
}
