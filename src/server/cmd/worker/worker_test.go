package main

import (
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	/* launch etcd */
	etcdCmd := []string{"docker", "run", "-p", "2379:2379",
	"docker pull quay.io/coreos/etcd:v3.1.1",

	// exec.Command(
	/* launch worker-shim */
}
