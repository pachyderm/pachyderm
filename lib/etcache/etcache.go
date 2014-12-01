package etcache

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"sync"
	"time"
)

var cache map[string]*etcd.Response = make(map[string]*etcd.Response)
var instertionTime map[string]time.Time = make(map[string]time.Time)
var lock sync.RWMutex

func Get(key string, sort, recursive bool) (*etcd.Response, error) {
	cacheKey := fmt.Sprint(key, "-", sort, "-", recursive)
	if time.Since(instertionTime[cacheKey]) > (5 * time.Minute) {
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
	if resp, err := client.Get(key, sort, recursive); err != nil {
		return resp, err
	} else {
		lock.Lock()
		defer lock.Unlock()
		cache[cacheKey] = resp
		instertionTime[cacheKey] = time.Now()
		return resp, nil
	}
}
