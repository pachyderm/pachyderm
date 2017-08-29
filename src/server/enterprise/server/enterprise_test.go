package enterprise

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestValidateActivationCode(t *testing.T) {
	_, err := validateActivationCode(testActivationCode)
	require.NoError(t, err)
}
