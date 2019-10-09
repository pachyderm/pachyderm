package work

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
)

const (
	etcdHost = "localhost"
	etcdPort = "32379"
)

// generateRandomString is a helper function for getPachClient
func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}

func TestWork(t *testing.T) {
	etcdPrefix := generateRandomString(32)
	taskID := "task"
	numSubtasks := 10
	numWorkers := 5
	// Setup etcd client and cleanup previous runs.
	etcdClient, err := getEtcdClient()
	require.NoError(t, err)
	// Setup maps for checking creation, processing, and collection.
	subtasksCreated := make(map[string]bool)
	subtasksProcessed := make(map[string]bool)
	var processMu sync.Mutex
	subtasksCollected := make(map[string]bool)
	// Setup workers.
	failProb := 0.25
	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				ctx, cancel := context.WithCancel(context.Background())
				w := NewWorker(etcdClient, etcdPrefix, func(_ context.Context, task, subtask *Task) error {
					if rand.Float64() < failProb {
						cancel()
					} else {
						processMu.Lock()
						subtasksProcessed[subtask.Id] = true
						processMu.Unlock()
					}
					return nil
				})
				if err := w.Run(ctx); err != nil {
					if ctx.Err() == context.Canceled {
						continue
					}
					break
				}
			}
		}()
	}
	// Setup master.
	m := NewMaster(etcdClient, etcdPrefix, func(_ context.Context, subtask *Task) error {
		subtasksCollected[subtask.Id] = true
		return nil
	})
	// Create task.
	var subtasks []*Task
	for i := 0; i < numSubtasks; i++ {
		id := strconv.Itoa(i)
		subtasks = append(subtasks, &Task{Id: id})
		subtasksCreated[id] = true
	}
	task := &Task{
		Id:       taskID,
		Subtasks: subtasks,
	}
	require.NoError(t, m.Run(context.Background(), task))
	require.Equal(t, subtasksCreated, subtasksProcessed, subtasksCollected)
}

func getEtcdClient() (*etcd.Client, error) {
	var etcdClient *etcd.Client
	if err := backoff.Retry(func() error {
		var err error
		etcdClient, err = etcd.New(etcd.Config{
			Endpoints:          []string{net.JoinHostPort(etcdHost, etcdPort)},
			DialOptions:        client.DefaultDialOptions(),
			MaxCallSendMsgSize: math.MaxInt32,
			MaxCallRecvMsgSize: math.MaxInt32,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize etcd client: %v", err)
		}
		return nil
	}, backoff.RetryEvery(time.Second).For(time.Minute)); err != nil {
		return nil, err
	}
	return etcdClient, nil
}
