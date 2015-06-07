package etcache

import (
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

var cache map[string]*etcd.Response = make(map[string]*etcd.Response)
var insertionTime map[string]time.Time = make(map[string]time.Time)
var lock sync.RWMutex

func Get(key string, sort, recursive bool) (*etcd.Response, error) {
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	if time.Since(insertionTime[cacheKey]) > (5 * time.Minute) {
		return ForceGet(key, sort, recursive)
	} else {
		lock.RLock()
		defer lock.RUnlock()
		return cache[cacheKey], nil
	}
}

func ForceGet(key string, sort, recursive bool) (*etcd.Response, error) {
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	resp, err := client.Get(key, sort, recursive)
	if err != nil {
		return resp, err
	}
	lock.Lock()
	defer lock.Unlock()
	cache[cacheKey] = resp
	insertionTime[cacheKey] = time.Now()
	return resp, nil
}

// Spoof artificially sets a value for a key that will never expire.
// This shouldn't be used outside of Tests.
func rawSpoof(key string, sort, recursive bool, val *etcd.Response) {
	lock.Lock()
	defer lock.Unlock()
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	cache[cacheKey] = val
	insertionTime[cacheKey] = time.Unix(1<<63-1, 999999999) // Maximal time
}

func Spoof1(key, val string) {
	resp := &etcd.Response{Node: &etcd.Node{Key: key, Value: val}}
	rawSpoof(key, false, false, resp)
}

func SpoofMany(key string, vals []string, sort bool) {
	var nodes []*etcd.Node
	for _, val := range vals {
		nodes = append(nodes, &etcd.Node{Value: val})
	}
	resp := &etcd.Response{Node: &etcd.Node{Key: key, Nodes: nodes}}
	rawSpoof(key, sort, true, resp)
}
