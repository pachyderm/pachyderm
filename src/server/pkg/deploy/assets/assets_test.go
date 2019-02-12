package assets

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestAddRegistry_NoSlash(t *testing.T) {
	imageName := "busybox"
	img := AddRegistry("", imageName)
	require.Equal(t, imageName, img)
}
