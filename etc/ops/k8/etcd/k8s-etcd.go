package main

/*
Copyright (c) 2015, Buoyant, Inc.
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

1. Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

import (
	"fmt"
	"net"
	"os"
	"os/exec"
)

func main() {
	for _, e := range os.Environ() {
		fmt.Printf("  %s\n", e)
	}

	name, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not determine hostname")
		os.Exit(1)
	}

	ip, err := getPodIP()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not determine IP address")
		os.Exit(1)
	}

	disco := ""
	if len(os.Args) >= 2 {
		disco = os.Args[1]
	}
	if disco == "" {
		fmt.Fprintln(os.Stderr, "Usage: k8s-etcd DISCOVERY")
		os.Exit(2)
	}

	clientURLs := fmt.Sprintf("http://%s:2379", ip)
	peerURLs := fmt.Sprintf("http://%s:2380", ip)

	fmt.Printf("$ etcd\n")
	fmt.Printf("    -name %s\n", name)
	fmt.Printf("    -discovery %s\n", disco)
	fmt.Printf("    -initial-cluster-state new\n")
	fmt.Printf("    -initial-advertise-peer-urls %s\n", peerURLs)
	fmt.Printf("    -listen-peer-urls %s\n", peerURLs)
	fmt.Printf("    -advertise-client-urls %s\n", clientURLs)
	fmt.Printf("    -listen-client-urls %s\n", clientURLs)

	cmd := exec.Command(
		"etcd",

		"-name", name,

		"-discovery", disco,

		"-initial-cluster-state", "new",

		"-initial-advertise-peer-urls", peerURLs,
		"-listen-peer-urls", peerURLs,

		"-advertise-client-urls", clientURLs,
		"-listen-client-urls", clientURLs,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(3)
	}
	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(4)
	}
	os.Exit(0)
}

// getPodIP guesses the local pod IP by returning the first IP address
// on an active, non-loopback network interface.
func getPodIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 1 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, err
			}

			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					return v.IP, nil
				case *net.IPAddr:
					return v.IP, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("Pod IP address could not be found on %d interfaces", len(ifaces))
}
