package assets

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestAddRegistry_NoSlash(t *testing.T) {
	imageName := "busybox"
	img := AddRegistry("", imageName)
	require.Equal(t, imageName, img)
}

func TestAddRegistry_OneSlash(t *testing.T) {
	imageName := "library/busybox"
	img := AddRegistry("", imageName)
	require.Equal(t, imageName, img)
}

func TestAddRegistry_WithRegistry(t *testing.T) {
	imageName := "library/busybox"
	registry := "example.com"
	expected := fmt.Sprintf("%s/%s", registry, imageName)
	img := AddRegistry(registry, imageName)
	require.Equal(t, expected, img)
}

func TestAddRegistry_SwitchRegistry(t *testing.T) {
	imageTag := "library/busybox"
	originalRegistry := "example.org"
	imageName := fmt.Sprintf("%s/%s", originalRegistry, imageTag)
	registry := "example.com"
	expected := fmt.Sprintf("%s/%s", registry, imageTag)
	img := AddRegistry(registry, imageName)
	require.Equal(t, expected, img)
}
