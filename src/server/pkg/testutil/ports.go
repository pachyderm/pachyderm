package testutil

import "sync/atomic"

var nextPort uint32 = 41300

// UniquePort generates a likely unique port, so that multiple servers can run
// concurrently. Note that it does not actually check that the port is free,
// but uses atomics and a fairly highly port range to maximize the likelihood
// that the port is available.
func UniquePort() uint16 {
	port := uint16(atomic.AddUint32(&nextPort, 1))

	if port == 0 {
		panic("ran out of ports!")
	}

	return port
}
