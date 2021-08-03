package testing

import (
	"bytes"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestLoad(t *testing.T) {
	env := testpachd.NewRealEnv(t, tu.NewTestDBConfig(t))
	c := env.PachClient
	resp, err := c.PfsAPIClient.RunLoadTestDefault(c.Ctx(), &types.Empty{})
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, cmdutil.Encoder("", buf).EncodeProto(resp))
	require.Equal(t, "", resp.Error, string(buf.Bytes()))
}
