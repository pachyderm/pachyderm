//go:build !livek8s
// +build !livek8s

package testutil

import (
	"testing"

	kube "k8s.io/client-go/kubernetes"
)

func GetKubeClient(t testing.TB) *kube.Clientset {
	t.Helper()
	t.Fatal("a kubernetes client can only be retrived when built with the 'livek8s' build tag")
	return nil
}
