//go:build unit_test

package testing

import (
	"bytes"
	"testing"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

func TestLoad(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient
	resp, err := c.PfsAPIClient.RunLoadTestDefault(c.Ctx(), &types.Empty{})
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, cmdutil.Encoder("", buf).EncodeProto(resp))
	require.Equal(t, "", resp.Error, buf.String())
}
