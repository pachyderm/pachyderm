package route

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.pachyderm.com/pachyderm/src/pkg/discovery"
	"go.pachyderm.com/pachyderm/src/pkg/require"
	"go.pachyderm.com/pachyderm/src/pkg/shard"
)

const (
	testNumShards   = 64
	testNumServers  = 8
	testNumReplicas = 3
)

func TestMasterOnly(t *testing.T) {
	client, err := getEtcdClient()
	require.NoError(t, err)
	runMasterOnlyTest(t, client)
}

func TestMasterReplica(t *testing.T) {
	client, err := getEtcdClient()
	require.NoError(t, err)
	runMasterReplicaTest(t, client)
}

type server struct {
	shards map[uint64]bool
	lock   sync.Mutex
	t      *testing.T
}

func (s *server) AddShard(shard uint64, version int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.shards[shard] = true
	return nil
}
func (s *server) RemoveShard(shard uint64, version int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.shards, shard)
	return nil
}

func (s *server) LocalShards() (map[uint64]bool, error) {
	return nil, nil
}

func newServer(t *testing.T) *server {
	return &server{make(map[uint64]bool), sync.Mutex{}, t}
}

type frontend struct {
	version int64
}

func (f *frontend) Version(version int64) error {
	f.version = version
	return nil
}

func newFrontend(t *testing.T) *frontend {
	return &frontend{shard.InvalidVersion}
}

type serverGroup struct {
	servers   []*server
	frontends []*frontend
	cancel    chan bool
	sharder   shard.Sharder
	offset    int
}

func NewServerGroup(t *testing.T, sharder shard.Sharder, numServers int, offset int) *serverGroup {
	serverGroup := serverGroup{cancel: make(chan bool), sharder: sharder, offset: offset}
	for i := 0; i < numServers; i++ {
		serverGroup.servers = append(serverGroup.servers, newServer(t))
	}
	return &serverGroup
}

func (s *serverGroup) run(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i, server := range s.servers {
		wg.Add(1)
		i := i
		server := server
		go func() {
			defer wg.Done()
			require.Equal(
				t,
				shard.ErrCancelled,
				s.sharder.Register(s.cancel, fmt.Sprintf("server-%d", i+s.offset), fmt.Sprintf("address-%d", i+s.offset), server),
			)
		}()
	}
	for i, frontend := range s.frontends {
		wg.Add(1)
		i := i
		frontend := frontend
		go func() {
			defer wg.Done()
			require.Equal(
				t,
				shard.ErrCancelled,
				s.sharder.RegisterFrontend(s.cancel, fmt.Sprintf("address-%d", i+s.offset), frontend),
			)
		}()
	}
}

func (s *serverGroup) satisfied(shardsLen int) bool {
	result := true
	for _, server := range s.servers {
		if len(server.shards) != shardsLen {
			result = false
		}
	}
	return result
}

func runMasterOnlyTest(t *testing.T, client discovery.Client) {
	sharder := shard.NewSharder(client, testNumShards, 0, "TestMasterOnly")
	cancel := make(chan bool)
	go func() {
		require.Equal(t, shard.ErrCancelled, sharder.AssignRoles(cancel))
	}()
	defer func() {
		close(cancel)
	}()
	serverGroup1 := NewServerGroup(t, sharder, testNumServers/2, 0)
	go serverGroup1.run(t)
	start := time.Now()
	for !serverGroup1.satisfied(testNumShards / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}
	serverGroup2 := NewServerGroup(t, sharder, testNumServers/2, testNumServers/2)
	go serverGroup2.run(t)
	start = time.Now()
	for !serverGroup1.satisfied(testNumShards/testNumServers) || !serverGroup2.satisfied(testNumShards/testNumServers) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}
	close(serverGroup1.cancel)
	start = time.Now()
	for !serverGroup2.satisfied(testNumShards / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(60) {
			t.Fatal("test timed out")
		}
	}
}

func runMasterReplicaTest(t *testing.T, client discovery.Client) {
	sharder := shard.NewSharder(client, testNumShards, testNumReplicas, "TestMasterReplica")
	cancel := make(chan bool)
	go func() {
		require.Equal(t, shard.ErrCancelled, sharder.AssignRoles(cancel))
	}()
	defer func() {
		close(cancel)
	}()
	serverGroup1 := NewServerGroup(t, sharder, testNumServers/2, 0)
	go serverGroup1.run(t)
	start := time.Now()
	for !serverGroup1.satisfied((testNumShards * (testNumReplicas + 1)) / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}

	serverGroup2 := NewServerGroup(t, sharder, testNumServers/2, testNumServers/2)
	go serverGroup2.run(t)
	start = time.Now()
	for !serverGroup1.satisfied((testNumShards*(testNumReplicas+1))/testNumServers) ||
		!serverGroup2.satisfied((testNumShards*(testNumReplicas+1))/testNumServers) {
		time.Sleep(time.Second)
		if time.Since(start) > time.Second*time.Duration(60) {
			t.Fatal("test timed out")
		}
	}

	close(serverGroup1.cancel)
	for !serverGroup2.satisfied((testNumShards * (testNumReplicas + 1)) / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(60) {
			t.Fatal("test timed out")
		}
	}
}

func getEtcdClient() (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
