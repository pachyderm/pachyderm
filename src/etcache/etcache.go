package etcache

import (
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

type Cache interface {
	Get(key string, sort bool, recursive bool) (*etcd.Response, error)
	ForceGet(key string, sort bool, recursive bool) (*etcd.Response, error)
}

type TestCache interface {
	Cache
	SpoofOne(key string, value string)
	SpoofMany(key string, values []string, sort bool)
}

func NewCache() Cache {
	return NewTestCache()
}

func NewTestCache() TestCache {
	return newTestCache()
}

type testCache struct {
	keyToResponse      map[string]*etcd.Response
	keyToInsertionTime map[string]time.Time
	lock               *sync.RWMutex
}

func newTestCache() *testCache {
	return &testCache{
		make(map[string]*etcd.Response),
		make(map[string]time.Time),
		&sync.RWMutex{},
	}
}

func (c *testCache) Get(key string, sort, recursive bool) (*etcd.Response, error) {
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	// Notice this works because the default value is 0
	if time.Since(c.keyToInsertionTime[cacheKey]) > (5 * time.Minute) {
		return c.ForceGet(key, sort, recursive)
	} else {
		c.lock.RLock()
		defer c.lock.RUnlock()
		return c.keyToResponse[cacheKey], nil
	}
}

func (c *testCache) ForceGet(key string, sort, recursive bool) (*etcd.Response, error) {
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	response, err := client.Get(key, sort, recursive)
	if err != nil {
		return response, err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.keyToResponse[cacheKey] = response
	c.keyToInsertionTime[cacheKey] = time.Now()
	return response, nil
}

func (c *testCache) SpoofOne(key string, value string) {
	response := &etcd.Response{Node: &etcd.Node{Key: key, Value: value}}
	c.rawSpoof(key, false, false, response)
}

func (c *testCache) SpoofMany(key string, values []string, sort bool) {
	var nodes []*etcd.Node
	for _, value := range values {
		nodes = append(nodes, &etcd.Node{Value: value})
	}
	response := &etcd.Response{Node: &etcd.Node{Key: key, Nodes: nodes}}
	c.rawSpoof(key, sort, true, response)
}

// Spoof artificially sets a value for a key that will never expire.
// This shouldn't be used outside of Tests.
func (c *testCache) rawSpoof(key string, sort bool, recursive bool, response *etcd.Response) {
	c.lock.Lock()
	defer c.lock.Unlock()
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	c.keyToResponse[cacheKey] = response
	c.keyToInsertionTime[cacheKey] = time.Now().Add(time.Hour * 100)
}
