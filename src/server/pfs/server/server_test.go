package server

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"math/rand"

	"golang.org/x/net/context"

	"go.pedge.io/proto/server"
	"google.golang.org/grpc"

	pclient "github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/shard"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	shards  = 32
	servers = 4

	ALPHABET = "abcdefghijklmnopqrstuvwxyz"
)

var (
	port int32 = 30651
)

func TestBlock(t *testing.T) {
	t.Parallel()
	blockClient := getBlockClient(t)
	_, err := blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "c1", 0),
		})
	require.NoError(t, err)
	_, err = blockClient.CreateDiff(
		context.Background(),
		&pfsclient.DiffInfo{
			Diff: pfsclient.NewDiff("foo", "c2", 0),
		})
	require.NoError(t, err)
	listDiffClient, err := blockClient.ListDiff(
		context.Background(),
		&pfsclient.ListDiffRequest{Shard: 0},
	)
	require.NoError(t, err)
	var diffInfos []*pfsclient.DiffInfo
	for {
		diffInfo, err := listDiffClient.Recv()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}
		diffInfos = append(diffInfos, diffInfo)
	}
	require.Equal(t, 3, len(diffInfos))
}

func TestSimple(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := "TestSimple"
	require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
	commit1, err := pfsutil.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))
	commitInfos, err := pfsclient.ListCommit(pfsClient, []string{repo})
	require.NoError(t, err)
	require.Equal(t, 1, len(commitInfos))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestBranch(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := "TestBranch"
	require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
	commit1, err := pfsutil.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, "master"))
	var buffer bytes.Buffer
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	branches, err := pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit1, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
	commit2, err := pfsclient.StartCommit(pfsClient, repo, "", "master")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, "master", "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, "master")
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, "master", "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
	branches, err = pfsclient.ListBranch(pfsClient, repo)
	require.NoError(t, err)
	require.Equal(t, commit2, branches[0].Commit)
	require.Equal(t, "master", branches[0].Branch)
}

func TestDisallowReadsDuringCommit(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)
	repo := "TestDisallowReadsDuringCommit"
	require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
	commit1, err := pfsutil.StartCommit(pfsClient, repo, "", "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit1.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)

	// Make sure we can't get the file before the commit is finished
	var buffer bytes.Buffer
	require.YesError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "", buffer.String())

	require.NoError(t, pfsclient.FinishCommit(pfsClient, repo, commit1.ID))
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	commit2, err := pfsclient.StartCommit(pfsClient, repo, commit1.ID, "")
	require.NoError(t, err)
	_, err = pfsclient.PutFile(pfsClient, repo, commit2.ID, "foo", 0, strings.NewReader("foo\n"))
	require.NoError(t, err)
	err = pfsclient.FinishCommit(pfsClient, repo, commit2.ID)
	require.NoError(t, err)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())

	// restart the server and make sure data is still there
	restartServer(server, t)
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit1.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\n", buffer.String())
	buffer = bytes.Buffer{}
	require.NoError(t, pfsclient.GetFile(pfsClient, repo, commit2.ID, "foo", 0, 0, "", nil, &buffer))
	require.Equal(t, "foo\nfoo\n", buffer.String())
}

func TestInspectRepoSimple(t *testing.T) {
	 t.Parallel()
	 pfsClient, _ := getClientAndServer(t)

	 repo := "TestInspectRepoSimple"
	 require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))

	 commit, err := pfsutil.StartCommit(pfsClient, repo, "", "")
	 require.NoError(t, err)

	 file1Content := "foo\n"
	 _, err = pfsutil.PutFile(pfsClient, repo, commit.ID, "foo", 0, strings.NewReader(file1Content))
	 require.NoError(t, err)

	 file2Content := "bar\n"
	 _, err = pfsutil.PutFile(pfsClient, repo, commit.ID, "bar", 0, strings.NewReader(file2Content))
	 require.NoError(t, err)

	 require.NoError(t, pfsutil.FinishCommit(pfsClient, repo, commit.ID))

	 info, err := pfsutil.InspectRepo(pfsClient, repo)
	 require.NoError(t, err)

	 require.Equal(t, int(info.SizeBytes), len(file1Content) + len(file2Content))
}

func TestInspectRepoComplex(t *testing.T) {
	 t.Parallel()
	 pfsClient, _ := getClientAndServer(t)

	 repo := "TestInspectRepoComplex"
	 require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))

	 commit, err := pfsutil.StartCommit(pfsClient, repo, "", "")
	 require.NoError(t, err)

	 numFiles := 100
	 minFileSize := 1000
	 maxFileSize := 2000
	 totalSize := 0

	 for i := 0; i < numFiles; i++ {
		 fileContent := generateRandomString(rand.Intn(maxFileSize-minFileSize)+minFileSize)
		 fileContent += "\n"
		 fileName := fmt.Sprintf("file_%d", i)
		 totalSize += len(fileContent)

		 _, err = pfsutil.PutFile(pfsClient, repo, commit.ID, fileName, 0, strings.NewReader(fileContent))
		 require.NoError(t, err)
	 }

	 require.NoError(t, pfsutil.FinishCommit(pfsClient, repo, commit.ID))

	 info, err := pfsutil.InspectRepo(pfsClient, repo)
	 require.NoError(t, err)

	 require.Equal(t, int(info.SizeBytes), totalSize)
}

func TestListRepo(t *testing.T) {
	t.Parallel()
	pfsClient, server := getClientAndServer(t)

	numRepos := 10
	repoNames := make(map[string]bool)
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
		repoNames[repo] = true
	}

	test := func() {
		repoInfos, err := pfsutil.ListRepo(pfsClient)
		require.NoError(t, err)

		for _, repoInfo := range repoInfos {
			require.True(t, repoNames[repoInfo.Repo.Name])
		}

		require.Equal(t, len(repoInfos), numRepos)
	}

	test()

	restartServer(server, t)

	test()
}

func TestDeleteRepo(t *testing.T) {
	t.Parallel()
	pfsClient, _ := getClientAndServer(t)

	numRepos := 10
	repoNames := make(map[string]bool)
	for i := 0; i < numRepos; i++ {
		repo := fmt.Sprintf("repo%d", i)
		require.NoError(t, pfsutil.CreateRepo(pfsClient, repo))
		repoNames[repo] = true
	}

	reposToRemove := 5
	for i := 0; i < reposToRemove; i++ {
		// Pick one random element from repoNames
		for repoName := range repoNames {
			require.NoError(t, pfsutil.DeleteRepo(pfsClient, repoName))
			delete(repoNames, repoName)
			break
		}
	}

	repoInfos, err := pfsutil.ListRepo(pfsClient)
	require.NoError(t, err)

	for _, repoInfo := range repoInfos {
		require.True(t, repoNames[repoInfo.Repo.Name])
	}

	require.Equal(t, len(repoInfos), numRepos-reposToRemove)
}

func generateRandomString(n int) string {
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
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pclient.Version},
			protoserver.ServeEnv{GRPCPort: uint16(localPort)},
		)
		require.NoError(t, err)
	}()
	<-ready
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	return pfsclient.NewBlockAPIClient(clientConn)
}

func runServers(t *testing.T, port int32, apiServer pfsclient.APIServer,
	internalAPIServer pfsclient.InternalAPIServer, blockAPIServer pfsclient.BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := protoserver.Serve(
			func(s *grpc.Server) {
				pfsclient.RegisterAPIServer(s, apiServer)
				pfsclient.RegisterInternalAPIServer(s, internalAPIServer)
				pfsclient.RegisterBlockAPIServer(s, blockAPIServer)
				close(ready)
			},
			protoserver.ServeOptions{Version: pclient.Version},
			protoserver.ServeEnv{GRPCPort: uint16(port)},
		)
		require.NoError(t, err)
	}()
	<-ready
}

func getClientAndServer(t *testing.T) (pfsclient.APIClient, []*internalAPIServer) {
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
	sharder := shard.NewLocalSharder(addresses, shards)
	var internalAPIServers []*internalAPIServer
	for i, port := range ports {
		address := addresses[i]
		driver, err := drive.NewDriver(address)
		require.NoError(t, err)
		blockAPIServer, err := NewLocalBlockAPIServer(root)
		require.NoError(t, err)
		hasher := pfsserver.NewHasher(shards, 1)
		dialer := grpcutil.NewDialer(grpc.WithInsecure())
		apiServer := NewAPIServer(hasher, shard.NewRouter(sharder, dialer, address))
		internalAPIServer := newInternalAPIServer(hasher, shard.NewRouter(sharder, dialer, address), driver)
		internalAPIServers = append(internalAPIServers, internalAPIServer)
		runServers(t, port, apiServer, internalAPIServer, blockAPIServer)
		for i := 0; i < shards; i++ {
			require.NoError(t, internalAPIServer.AddShard(uint64(i)))
		}
	}
	clientConn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	require.NoError(t, err)
	return pfsclient.NewAPIClient(clientConn), internalAPIServers
}

func restartServer(servers []*internalAPIServer, t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, server := range servers {
		server := server
		for i := 0; i < shards; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				require.NoError(t, server.DeleteShard(uint64(i)))
				require.NoError(t, server.AddShard(uint64(i)))
			}()
		}
	}
}

func uniqueString(prefix string) string {
	return prefix + "." + uuid.NewWithoutDashes()[0:12]
}
