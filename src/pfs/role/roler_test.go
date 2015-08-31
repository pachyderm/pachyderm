package role

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/stretchr/testify/require"
)

const (
	testNumShards   = 64
	testNumServers  = 8
	testNumReplicas = 3
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
	roles map[int]string
	t     *testing.T
}

func (s *server) Master(shard int) error {
	role, ok := s.roles[shard]
	require.False(s.t, ok, "Master %d assigned when we're already %s for it.", shard, role)
	s.roles[shard] = "master"
	return nil
}
func (s *server) Replica(shard int) error {
	role, ok := s.roles[shard]
	require.False(s.t, ok, "Replica %d assigned when we're already %s for it.", shard, role)
	s.roles[shard] = "replica"
	return nil
}
func (s *server) Clear(shard int) error {
	delete(s.roles, shard)
	return nil
}

func newServer(t *testing.T) *server {
	return &server{make(map[int]string), t}
}

type serverGroup struct {
	servers []*server
	rolers  []Roler
	offset  int
}

func NewServerGroup(t *testing.T, addresser route.Addresser, numServers int, offset int, numReplicas int) *serverGroup {
	sharder := route.NewSharder(testNumShards)
	serverGroup := serverGroup{offset: offset}
	for i := 0; i < numServers; i++ {
		serverGroup.servers = append(serverGroup.servers, newServer(t))
		serverGroup.rolers = append(serverGroup.rolers, NewRoler(addresser, sharder, serverGroup.servers[i], fmt.Sprintf("server-%d", i+offset), numReplicas))
	}
	return &serverGroup
}

func (s *serverGroup) run(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, roler := range s.rolers {
		wg.Add(1)
		go func(roler Roler) {
			defer wg.Done()
			require.Equal(t, etcd.ErrWatchStoppedByUser, roler.Run())
		}(roler)
	}
}

func (s *serverGroup) cancel() {
	for _, roler := range s.rolers {
		roler.Cancel()
	}
}

func (s *serverGroup) satisfied(rolesLen int) bool {
	result := true
	for _, server := range s.servers {
		if len(server.roles) != rolesLen {
			result = false
		}
	}
	return result
}

func runMasterOnlyTest(t *testing.T, client discovery.Client) {
	addresser := route.NewDiscoveryAddresser(client, "TestMasterOnlyRoler")
	serverGroup1 := NewServerGroup(t, addresser, testNumServers/2, 0, 0)
	go serverGroup1.run(t)
	start := time.Now()
	for !serverGroup1.satisfied(testNumShards / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}

	serverGroup2 := NewServerGroup(t, addresser, testNumServers/2, testNumServers/2, 0)
	go serverGroup2.run(t)
	start = time.Now()
	for !serverGroup1.satisfied(testNumShards/testNumServers) || !serverGroup2.satisfied(testNumShards/testNumServers) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}

	serverGroup1.cancel()
	for !serverGroup2.satisfied(testNumShards / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(60) {
			t.Fatal("test timed out")
		}
	}
}

func runMasterReplicaTest(t *testing.T, client discovery.Client) {
	addresser := route.NewDiscoveryAddresser(client, "TestMasterReplicaRoler")
	serverGroup1 := NewServerGroup(t, addresser, testNumServers/2, 0, testNumReplicas)
	go serverGroup1.run(t)
	start := time.Now()
	for !serverGroup1.satisfied((testNumShards * (testNumReplicas + 1)) / (testNumServers / 2)) {
		time.Sleep(500 * time.Millisecond)
		if time.Since(start) > time.Second*time.Duration(30) {
			t.Fatal("test timed out")
		}
	}

	serverGroup2 := NewServerGroup(t, addresser, testNumServers/2, testNumServers/2, testNumReplicas)
	go serverGroup2.run(t)
	start = time.Now()
	for !serverGroup1.satisfied((testNumShards*(testNumReplicas+1))/testNumServers) ||
		!serverGroup2.satisfied((testNumShards*(testNumReplicas+1))/testNumServers) {
		time.Sleep(time.Second)
		if time.Since(start) > time.Second*time.Duration(60) {
			t.Fatal("test timed out")
		}
	}

	serverGroup1.cancel()
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
