package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestValidateActivationCode(t *testing.T) {
	_, err := license.Validate(testutil.GetTestEnterpriseCode(t))
	require.NoError(t, err)
}
