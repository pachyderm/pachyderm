package testing

import (
	"bytes"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestLoad(t *testing.T) {
	pachClient := pachd.NewTestPachd(t)
	resp, err := pachClient.DebugClient.RunPFSLoadTestDefault(pachClient.Ctx(), &emptypb.Empty{})
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, cmdutil.Encoder("", buf).EncodeProto(resp))
	require.Equal(t, "", resp.Error, buf.String())
}
