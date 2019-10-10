package grpcutil

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	// DefaultPachdNodePort is the pachd kubernetes service's default
	// NodePort.Port setting
	DefaultPachdNodePort = "30650"

	// DefaultPachdPort is the pachd kubernetes service's default
	// Port (often used with Pachyderm ELBs)
	DefaultPachdPort = "650"
)

// SanitizePachAddress validates and sanitizes a pachd address.
func SanitizePachAddress(address string) (string, error) {
	// allow no specification of a pachd address
	if address == "" {
		return "", nil
	}

	// parse
	u, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("could not parse pachd address: %v", err)
	}

	// validate
	if u.Scheme != "" && u.Scheme != "grpc" && u.Scheme != "grpcs" {
		return "", fmt.Errorf("unrecognized scheme in pachd address: %s", u.Scheme)
	}
	if u.User != nil {
		return "", errors.New("pachd address should not include login credentials")
	}
	if u.Path != "" {
		return "", errors.New("pachd address should not include a path")
	}
	if u.RawQuery != "" {
		return "", errors.New("pachd address should not include a query string")
	}
	if u.Fragment != "" {
		return "", errors.New("pachd address should not include a fragment")
	}

	// sanitize
	if u.Scheme == "" {
		u.Scheme = "grpc"
	}
	if !strings.Contains(u.Host, ":") {
		u.Host = fmt.Sprintf("%s:%d", u.Host, DefaultPachdNodePort)
	}

	return u.String(), nil
}

// IsTLSPachdAddress returns whether a given address has TLS enabled
func IsTLSPachdAddress(address string) bool {
	if address == "" {
		return false
	}
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	return u.Scheme == "grpcs"
}
