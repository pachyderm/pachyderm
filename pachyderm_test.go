package pachyderm

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"google.golang.org/grpc"
)

func TestSimple(t *testing.T) {
	pfsClient := getPfsClient(t)
	require.NoError(t, pfsutil.CreateRepo(pfsClient, "data"))
	require.NoError(t, pfsutil.CreateRepo(pfsClient, "output"))
	commit, err := pfsutil.StartCommit(pfsClient, "data", "")
	require.NoError(t, err)
	_, err = pfsutil.PutFile(pfsClient, "data", commit.Id, "file", 0, strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, pfsutil.FinishCommit(pfsClient, "data", commit.Id))
	ppsClient := getPpsClient(t)
	_, err = ppsutil.CreateJob(ppsClient, "", []string{"cp", "/pfs/data/file", "/pfs/output/file"}, 1, []*pfs.Commit{commit}, nil)
	require.NoError(t, err)
}

func getPfsClient(t *testing.T) pfs.APIClient {
	pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR")
	if pfsdAddr == "" {
		t.Error("PFSD_PORT_650_TCP_ADDR not set")
	}
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:650", pfsdAddr), grpc.WithInsecure())
	require.NoError(t, err)
	return pfs.NewAPIClient(clientConn)
}

func getPpsClient(t *testing.T) pps.APIClient {
	ppsdAddr := os.Getenv("PPSD_PORT_651_TCP_ADDR")
	if ppsdAddr == "" {
		t.Error("PPSD_PORT_651_TCP_ADDR not set")
	}
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:651", ppsdAddr), grpc.WithInsecure())
	require.NoError(t, err)
	return pps.NewAPIClient(clientConn)
}
