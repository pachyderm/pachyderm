package server

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	pclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
)

const (
	servers = 2

	ALPHABET = "abcdefghijklmnopqrstuvwxyz"
)

var (
	port int32 = 30651
)

var testDBs []string

func TestInvalidRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)
	require.YesError(t, client.CreateRepo("/repo"))

	require.NoError(t, client.CreateRepo("lenny"))
	require.NoError(t, client.CreateRepo("lenny123"))
	require.NoError(t, client.CreateRepo("lenny_123"))

	require.YesError(t, client.CreateRepo("lenny-123"))
	require.YesError(t, client.CreateRepo("lenny.123"))
	require.YesError(t, client.CreateRepo("lenny:"))
	require.YesError(t, client.CreateRepo("lenny,"))
	require.YesError(t, client.CreateRepo("lenny#"))
}

func TestCreateRepoNonexistantProvenance(t *testing.T) {
	// This method of calling CreateRepo
	// is used within pps CreateJob()

	client := getClient(t)
	var provenance []*pfs.Repo
	provenance = append(provenance, pclient.NewRepo("bogusABC"))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo("foo"),
			Provenance: provenance,
		},
	)
	require.YesError(t, err)
}

func TestInspectRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	repo := "repo"
	require.NoError(t, client.CreateRepo(repo))

	repoInfo, err := client.InspectRepo(repo)
	require.NoError(t, err)
	require.Equal(t, repo, repoInfo.Repo.Name)
	require.NotNil(t, repoInfo.Created)
	require.Equal(t, 0, int(repoInfo.SizeBytes))
}

func TestListRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numRepos := 10
	var repoNames []string
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, client.CreateRepo(repo))
		repoNames = append(repoNames, repo)
	}

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)

	for i, repoInfo := range repoInfos {
		require.Equal(t, repoNames[i], repoInfo.Repo.Name)
	}

	require.Equal(t, len(repoInfos), numRepos)
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	numRepos := 10
	repoNames := make(map[string]bool)
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, client.CreateRepo(repo))
		repoNames[repo] = true
	}

	reposToRemove := 5
	for i := 0; i < reposToRemove; i++ {
		// Pick one random element from repoNames
		for repoName := range repoNames {
			require.NoError(t, client.DeleteRepo(repoName, false))
			delete(repoNames, repoName)
			break
		}
	}

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)

	for _, repoInfo := range repoInfos {
		require.True(t, repoNames[repoInfo.Repo.Name])
	}

	require.Equal(t, len(repoInfos), numRepos-reposToRemove)
}

func TestDeleteProvenanceRepo(t *testing.T) {
	t.Parallel()
	client := getClient(t)

	// Create two repos, one as another's provenance
	require.NoError(t, client.CreateRepo("A"))
	_, err := client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfs.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)

	// Delete the provenance repo; that should fail.
	require.YesError(t, client.DeleteRepo("A", false))

	// Delete the leaf repo, then the provenance repo; that should succeed
	require.NoError(t, client.DeleteRepo("B", false))
	require.NoError(t, client.DeleteRepo("A", false))

	repoInfos, err := client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 0, len(repoInfos))

	// Create two repos again
	require.NoError(t, client.CreateRepo("A"))
	_, err = client.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo:       pclient.NewRepo("B"),
			Provenance: []*pfs.Repo{pclient.NewRepo("A")},
		},
	)
	require.NoError(t, err)

	// Force delete should succeed
	require.NoError(t, client.DeleteRepo("A", true))

	repoInfos, err = client.ListRepo(nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(repoInfos))
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = ALPHABET[rand.Intn(len(ALPHABET))]
	}
	return string(b)
}

func getBlockClient(t *testing.T) pfs.BlockAPIClient {
	localPort := atomic.AddInt32(&port, 1)
	address := fmt.Sprintf("localhost:%d", localPort)
	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	blockAPIServer, err := NewLocalBlockAPIServer(root)
	require.NoError(t, err)
	ready := make(chan bool)
	go func() {
		err := grpcutil.Serve(
			func(s *grpc.Server) {
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			grpcutil.ServeOptions{Version: version.Version},
			grpcutil.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfs.NewBlockAPIClient(clientConn)
}

func runServers(t *testing.T, port int32, apiServer pfs.APIServer,
	blockAPIServer pfs.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := grpcutil.Serve(
			func(s *grpc.Server) {
				pfs.RegisterAPIServer(s, apiServer)
				pfs.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			grpcutil.ServeOptions{
				Version:    version.Version,
				MaxMsgSize: MaxMsgSize,
			},
			grpcutil.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func getClient(t *testing.T) pclient.APIClient {
	dbName := "pachyderm_test_" + uuid.NewWithoutDashes()[0:12]
	testDBs = append(testDBs, dbName)

	root := uniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	var ports []int32
	for i := 0; i < servers; i++ {
		ports = append(ports, atomic.AddInt32(&port, 1))
	}
	var addresses []string
	for _, port := range ports {
		addresses = append(addresses, fmt.Sprintf("localhost:%d", port))
	}
	prefix := generateRandomString(32)
	for i, port := range ports {
		address := addresses[i]
		driver, err := drive.NewLocalDriver(address, prefix)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		apiServer := newAPIServer(driver, nil)
		runServers(t, port, apiServer, blockAPIServer)
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pclient.APIClient{PfsAPIClient: pfs.NewAPIClient(clientConn)}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
