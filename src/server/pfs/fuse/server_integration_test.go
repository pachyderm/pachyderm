//go:build k8s

package fuse

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	c = tu.AuthenticateClient(t, c, auth.RootUser)

	withServerMount(t, c, nil, func(mountPoint string) {
		type Config struct {
			ClusterStatus string `json:"cluster_status"`
			PachdAddress  string `json:"pachd_address"`
		}

		// PUT
		invalidCfg := &Config{ClusterStatus: "INVALID", PachdAddress: "bad_address"}
		m := map[string]string{"pachd_address": invalidCfg.PachdAddress}
		b := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(m))

		putResp, err := put("config", b)
		require.NoError(t, err)
		require.Equal(t, 500, putResp.StatusCode)

		cfg := &Config{ClusterStatus: "AUTH_ENABLED", PachdAddress: c.GetAddress().Qualified()}
		m = map[string]string{"pachd_address": cfg.PachdAddress}
		b = new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(b).Encode(m))

		putResp, err = put("config", b)
		require.NoError(t, err)
		defer putResp.Body.Close()

		putConfig := &Config{}
		require.NoError(t, json.NewDecoder(putResp.Body).Decode(putConfig))

		cfgParsedPachdAddress, err := grpcutil.ParsePachdAddress(cfg.PachdAddress)
		require.NoError(t, err)

		require.Equal(t, cfg.ClusterStatus, putConfig.ClusterStatus)
		require.Equal(t, cfgParsedPachdAddress.Qualified(), putConfig.PachdAddress)
		require.Equal(t, cfgParsedPachdAddress.Qualified(), c.GetAddress().Qualified())

		// GET
		getResp, err := get("config")
		require.NoError(t, err)
		defer getResp.Body.Close()

		getConfig := &Config{}
		require.NoError(t, json.NewDecoder(getResp.Body).Decode(getConfig))

		require.Equal(t, cfg.ClusterStatus, getConfig.ClusterStatus)
		require.Equal(t, cfg.PachdAddress, getConfig.PachdAddress)
	})
}

func TestRepoAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	alice, bob := "robot"+tu.UniqueString("alice"), "robot"+tu.UniqueString("bob")
	aliceClient, bobClient := tu.AuthenticateClient(t, c, alice), tu.AuthenticateClient(t, c, bob)

	require.NoError(t, aliceClient.CreateRepo("repo1"))
	commit := client.NewCommit("repo1", "master", "")
	err := aliceClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)

	withServerMount(t, aliceClient, nil, func(mountPoint string) {
		resp, err := get("repos")
		require.NoError(t, err)

		reposResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(reposResp))
		require.Equal(t, (*reposResp)["repo1"].Authorization, "write")
	})

	withServerMount(t, bobClient, nil, func(mountPoint string) {
		resp, err := get("repos")
		require.NoError(t, err)

		reposResp := &ListRepoResponse{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(reposResp))
		require.Equal(t, (*reposResp)["repo1"].Authorization, "none")

		resp, _ = put("repos/repo1/master/_mount?name=repo1&mode=ro", nil)
		require.Equal(t, resp.StatusCode, 500)
	})
}

//TODO(fahad): requires identity server to work with realEnv.
func TestUnauthenticatedCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	withServerMount(t, c, nil, func(mountPoint string) {
		resp, _ := get("repos")
		require.Equal(t, 200, resp.StatusCode)
	})

	c = tu.UnauthenticatedPachClient(t, c)
	withServerMount(t, c, nil, func(mountPoint string) {
		resp, _ := get("repos")
		require.Equal(t, 401, resp.StatusCode)
	})

	c = tu.AuthenticateClient(t, c, "test")
	withServerMount(t, c, nil, func(mountPoint string) {
		resp, _ := get("repos")
		require.Equal(t, 200, resp.StatusCode)
	})
}

//TODO(fahad): requires identity server to work with realEnv.
func TestAuthLoginLogout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	tu.ActivateAuthClient(t, c)
	require.NoError(t, tu.ConfigureOIDCProvider(t, c))
	c = tu.UnauthenticatedPachClient(t, c)

	withServerMount(t, c, nil, func(mountPoint string) {
		authResp, err := put("auth/_login", nil)
		require.NoError(t, err)
		defer authResp.Body.Close()

		type AuthLoginResp struct {
			AuthUrl string `json:"auth_url"`
		}
		getAuthLogin := &AuthLoginResp{}
		require.NoError(t, json.NewDecoder(authResp.Body).Decode(getAuthLogin))

		tu.DoOAuthExchange(t, c, c, getAuthLogin.AuthUrl)
		time.Sleep(1 * time.Second)

		_, err = c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.NoError(t, err)

		_, err = put("auth/_logout", nil)
		require.NoError(t, err)

		_, err = c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
		require.ErrorIs(t, err, auth.ErrNotSignedIn)
	})
}
