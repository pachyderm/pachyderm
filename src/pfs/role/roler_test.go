package role

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/stretchr/testify/require"
)

const (
	testNumShards  = 2
	testNumServers = 2
)

func TestRoler(t *testing.T) {
	client, err := getEtcdClient()
	require.NoError(t, err)
	runTest(t, client)
}

type server struct {
	roles map[int]string
}

func (s *server) Master(shard int) error {
	s.roles[shard] = "master"
	return nil
}
func (s *server) Replica(shard int) error {
	s.roles[shard] = "replica"
	return nil
}
func (s *server) Clear(shard int) error {
	log.Print("Clear.")
	delete(s.roles, shard)
	return nil
}

func newServer() *server {
	return &server{make(map[int]string)}
}

type serverGroup struct {
	servers []*server
	rolers  []Roler
}

func NewServerGroup(addresser route.Addresser, numServers int, offset int) *serverGroup {
	sharder := route.NewSharder(testNumShards)
	serverGroup := serverGroup{}
	for i := 0; i < numServers; i++ {
		serverGroup.servers = append(serverGroup.servers, newServer())
		serverGroup.rolers = append(serverGroup.rolers, NewRoler(addresser, sharder, serverGroup.servers[i], fmt.Sprintf("server-%d", i+offset)))
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
			require.NoError(t, roler.Run())
		}(roler)
	}
}

func (s *serverGroup) satisfied(rolesLen int) bool {
	for _, server := range s.servers {
		if len(server.roles) != rolesLen {
			log.Printf("len(server.roles): %d, rolesLen: %d", len(server.roles), rolesLen)
			return false
		}
	}
	return true
}

func runTest(t *testing.T, client discovery.Client) {
	addresser := route.NewDiscoveryAddresser(client, "TestRoler")
	serverGroup := NewServerGroup(addresser, testNumServers/2, 0)
	go serverGroup.run(t)
	start := time.Now()
	for !serverGroup.satisfied(testNumShards / (testNumServers / 2)) {
		time.Sleep(3 * time.Second)
		if time.Since(start) > time.Second*time.Duration(10) {
			t.Fatal("test timed out")
		}
	}

	serverGroup2 := NewServerGroup(addresser, testNumServers/2, testNumServers/2)
	go serverGroup2.run(t)
	start = time.Now()
	for !serverGroup.satisfied(testNumShards/testNumServers) || !serverGroup2.satisfied(testNumShards/testNumServers) {
		time.Sleep(3 * time.Second)
		if time.Since(start) > time.Second*time.Duration(10) {
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
