//go:build !livek8s
// +build !livek8s

package testutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

func GetPachClient(t testing.TB) *client.APIClient {
	t.Helper()
	t.Fatal("a pachyderm client can only be retrieved when built with the 'livek8s' build tag")
	return nil
}
