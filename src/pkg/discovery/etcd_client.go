package discovery

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

var (
	holdTTL uint64 = 20
)

type etcdClient struct {
	client *etcd.Client
}

func newEtcdClient(addresses ...string) *etcdClient {
	return &etcdClient{etcd.NewClient(addresses)}
}

func (c *etcdClient) Close() error {
	c.client.Close()
	return nil
}

func (c *etcdClient) Get(key string) (string, bool, error) {
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			return "", false, nil
		}
		return "", false, err
	}
	return response.Node.Value, true, nil
}

func (c *etcdClient) GetAll(key string) (map[string]string, error) {
	response, err := c.client.Get(key, false, true)
	result := make(map[string]string, 0)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			return result, nil
		}
		return nil, err
	}
	nodeToMap(response.Node, result)
	return result, nil
}

func (c *etcdClient) Watch(key string, cancel chan bool, callBack func(string) error) error {
	var waitIndex uint64 = 1
	// First get the starting value of the key
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			if err := callBack(""); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err := callBack(response.Node.Value); err != nil {
			return err
		}
		waitIndex = response.Node.ModifiedIndex + 1
	}
	for {
		response, err := c.client.Watch(key, waitIndex, false, nil, cancel)
		if err != nil {
			return err
		}
		if err := callBack(response.Node.Value); err != nil {
			return err
		}
		waitIndex = response.Node.ModifiedIndex + 1
	}
}

func (c *etcdClient) WatchAll(key string, cancel chan bool, callBack func(map[string]string) error) (retErr error) {
	var waitIndex uint64 = 1
	value := make(map[string]string)
	// First get the starting value of the key
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			if err := callBack(nil); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if nodeToMap(response.Node, value) {
			log.Print("starter value: ", value)
			if err := callBack(value); err != nil {
				return err
			}
		}
		waitIndex = maxModifiedIndex(response.Node) + 1
		log.Print("starter waitIndex: ", waitIndex)

	}
	for {
		response, err := c.client.Watch(key, waitIndex, true, nil, cancel)
		if err != nil {
			return err
		}
		log.Printf("response.Node: %+v", response.Node)
		if nodeToMap(response.Node, value) {
			log.Print("watch value ", value)
			if err := callBack(value); err != nil {
				return err
			}
		}
		waitIndex = maxModifiedIndex(response.Node) + 1
		log.Print("watch  waitIndex: ", waitIndex)
	}
}

func (c *etcdClient) Set(key string, value string, ttl uint64) error {
	log.Printf("set %s", key)
	_, err := c.client.Set(key, value, ttl)
	return err
}

func (c *etcdClient) Create(key string, value string, ttl uint64) error {
	_, err := c.client.Create(key, value, ttl)
	return err
}

func (c *etcdClient) CreateInDir(dir string, value string, ttl uint64) error {
	_, err := c.client.CreateInOrder(dir, value, ttl)
	return err
}

func (c *etcdClient) Delete(key string) error {
	_, err := c.client.Delete(key, false)
	return err
}

func (c *etcdClient) CheckAndSet(key string, value string, ttl uint64, oldValue string) error {
	_, err := c.client.CompareAndSwap(key, value, ttl, oldValue, 0)
	return err
}

func (c *etcdClient) Hold(key string, value string, oldValue string, cancel chan bool) error {
	for {
		var err error
		if oldValue == "" {
			_, err = c.client.Create(key, value, holdTTL)
		} else {
			_, err = c.client.CompareAndSwap(key, value, holdTTL, oldValue, 0)
		}
		if err != nil {
			return err
		}
		oldValue = value
		cancel := make(chan bool)
		time.AfterFunc(time.Second*time.Duration(holdTTL/2), func() { close(cancel) })
		if err := c.Watch(key, cancel, func(newValue string) error {
			if newValue != value {
				return fmt.Errorf("pachyderm: lost hold")
			}
			return nil
		}); err != nil && err != etcd.ErrWatchStoppedByUser {
			return err
		}
	}
}

// nodeToMap translates the contents of a node into a map
// nodeToMap can be called on the same map with successive results from watch
// to accumulate a value
// nodeToMap returns true if out was modified
func nodeToMap(node *etcd.Node, out map[string]string) bool {
	key := strings.TrimPrefix(node.Key, "/")
	if !node.Dir {
		if node.Value == "" {
			if _, ok := out[key]; ok {
				delete(out, key)
				return true
			}
			return false
		}
		if value, ok := out[key]; !ok || value != node.Value {
			out[key] = node.Value
			return true
		}
		return false
	}
	changed := false
	for _, node := range node.Nodes {
		changed = nodeToMap(node, out) || changed
	}
	return changed
}

func maxModifiedIndex(node *etcd.Node) uint64 {
	result := node.ModifiedIndex
	for _, node := range node.Nodes {
		if modifiedIndex := maxModifiedIndex(node); modifiedIndex > result {
			result = modifiedIndex
		}
	}
	return result
}
