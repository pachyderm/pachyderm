package route

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/require"
)

const (
	testNumShards   = 512
	testNumServers  = 8
	testNumReplicas = 1
)

func TestMasterOnlyRoler(t *testing.T) {
	client, err := getEtcdClient()
	require.NoError(t, err)
	runMasterOnlyTest(t, client)
}

func TestMasterReplicaRoler(t *testing.T) {
	t.Skip()
	client, err := getEtcdClient()
	require.NoError(t, err)
	runMasterReplicaTest(t, client)
}

type server struct {
	shards map[uint64]bool
	lock   sync.Mutex
	t      *testing.T
}

func (s *server) AddShard(shard uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.shards[shard] = true
	return nil
}
func (s *server) RemoveShard(shard uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.shards, shard)
	return nil
}

func (s *server) LocalShards() ([]uint64, error) {
	return nil, nil
}

func newServer(t *testing.T) *server {
	return &server{make(map[uint64]bool), sync.Mutex{}, t}
}

type serverGroup struct {
	servers   []*server
	cancel    chan bool
	addresser Addresser
	offset    int
}

func NewServerGroup(t *testing.T, addresser Addresser, numServers int, offset int) *serverGroup {
	serverGroup := serverGroup{cancel: make(chan bool), addresser: addresser, offset: offset}
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
		go func(i int, server Server) {
			defer wg.Done()
			require.Equal(
				t,
				ErrCancelled,
				s.addresser.Register(s.cancel, fmt.Sprintf("server-%d", i+s.offset), fmt.Sprintf("address-%d", i+s.offset), server),
			)
		}(i, server)
	}
}

func (s *serverGroup) satisfied(shardsLen int) bool {
	result := true
	for _, server := range s.servers {
		if len(server.shards) != shardsLen {
			log.Printf("have: %d, want: %d", len(server.shards), shardsLen)
			result = false
		}
	}
	return result
}

func runMasterOnlyTest(t *testing.T, client discovery.Client) {
	sharder := NewSharder(testNumShards, 0)
	addresser := NewDiscoveryAddresser(client, sharder, "TestMasterOnlyRoler")
	cancel := make(chan bool)
	go func() {
		require.Equal(t, ErrCancelled, addresser.AssignRoles(cancel))
	}()
	defer func() {
		close(cancel)
	}()
	log.Print("Start Group 1")
	serverGroup1 := NewServerGroup(t, addresser, testNumServers/2, 0)
	go serverGroup1.run(t)
	start := time.Now()
	for !serverGroup1.satisfied(testNumShards / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}
	log.Print("Start Group 2")
	serverGroup2 := NewServerGroup(t, addresser, testNumServers/2, testNumServers/2)
	go serverGroup2.run(t)
	start = time.Now()
	for !serverGroup1.satisfied(testNumShards/testNumServers) || !serverGroup2.satisfied(testNumShards/testNumServers) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}
	log.Print("Cancel Group 1")
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
	sharder := NewSharder(testNumShards, testNumReplicas)
	addresser := NewDiscoveryAddresser(client, sharder, "TestMasterReplicaRoler")
	cancel := make(chan bool)
	go addresser.AssignRoles(cancel)
	defer func() {
		close(cancel)
	}()
	serverGroup1 := NewServerGroup(t, addresser, testNumServers/2, 0)
	go serverGroup1.run(t)
	start := time.Now()
	for !serverGroup1.satisfied((testNumShards * (testNumReplicas + 1)) / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}

	serverGroup2 := NewServerGroup(t, addresser, testNumServers/2, testNumServers/2)
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
