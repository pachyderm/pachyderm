package main

import (
	"bytes"
	"strings"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	inputRepo      string = "input"
	outputRepo     string = "output"
	upgradeSubject string = "upgrade_client"
)

func TestPreUpgrade(t *testing.T) {
	c := tu.GetAuthenticatedPachClient(t, upgradeSubject)
	require.NoError(t, c.CreateRepo(inputRepo))
	require.NoError(t,
		c.CreatePipeline(outputRepo,
			"busybox",
			[]string{"sh"},
			[]string{"cp /pfs/input/* /pfs/out/;"},
			nil,
			&pps.Input{Pfs: &pps.PFSInput{Glob: "/", Repo: inputRepo}},
			"master",
			false,
		))
	require.NoError(t, c.WithModifyFileClient(client.NewCommit(inputRepo, "master", ""), func(mf client.ModifyFile) error {
		return errors.EnsureStack(mf.PutFile("foo", strings.NewReader("foo")))
	}))

	commitInfo, err := c.InspectCommit(outputRepo, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
	require.NoError(t, err)

	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Branch.Repo, client.NewRepo(outputRepo)) {
			require.NoError(t, c.GetFile(info.Commit, "foo", &buf))
			require.Equal(t, "foo", buf.String())
		}
	}
}

func TestPostUpgrade(t *testing.T) {
	c := tu.GetAuthenticatedPachClient(t, upgradeSubject)

	state, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_ACTIVE, state.State)

	require.NoError(t, c.WithModifyFileClient(client.NewCommit(inputRepo, "master", ""), func(mf client.ModifyFile) error {
		return errors.EnsureStack(mf.PutFile("bar", strings.NewReader("bar")))
	}))

	commitInfo, err := c.InspectCommit(outputRepo, "master", "")
	require.NoError(t, err)
	commitInfos, err := c.WaitCommitSetAll(commitInfo.Commit.ID)
	require.NoError(t, err)
	var buf bytes.Buffer
	for _, info := range commitInfos {
		if proto.Equal(info.Commit.Branch.Repo, client.NewRepo(outputRepo)) {
			require.NoError(t, c.GetFile(info.Commit, "foo", &buf))
			require.Equal(t, "foo", buf.String())

			buf.Reset()

			require.NoError(t, c.GetFile(info.Commit, "bar", &buf))
			require.Equal(t, "bar", buf.String())
		}
	}
}
