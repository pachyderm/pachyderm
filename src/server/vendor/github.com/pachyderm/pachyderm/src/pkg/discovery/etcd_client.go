package discovery

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
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

func (c *etcdClient) Get(key string) (string, error) {
	response, err := c.client.Get(key, false, false)
	if err != nil {
		return "", err
	}
	return response.Node.Value, nil
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
	// This retry is needed for when the etcd cluster gets overloaded.
	for {
		if err := c.watchWithoutRetry(key, cancel, callBack); err != nil {
			etcdErr, ok := err.(*etcd.EtcdError)
			if ok && etcdErr.ErrorCode == 401 {
				continue
			}
			if ok && etcdErr.ErrorCode == 501 {
				continue
			}
			return err
		}
	}
}

func (c *etcdClient) WatchAll(key string, cancel chan bool, callBack func(map[string]string) error) error {
	for {
		if err := c.watchAllWithoutRetry(key, cancel, callBack); err != nil {
			etcdErr, ok := err.(*etcd.EtcdError)
			if ok && etcdErr.ErrorCode == 401 {
				continue
			}
			if ok && etcdErr.ErrorCode == 501 {
				continue
			}
			return err
		}
	}
}

func (c *etcdClient) Set(key string, value string, ttl uint64) error {
	_, err := c.client.Set(key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) Create(key string, value string, ttl uint64) error {
	_, err := c.client.Create(key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) CreateInDir(dir string, value string, ttl uint64) error {
	_, err := c.client.CreateInOrder(dir, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) Delete(key string) error {
	_, err := c.client.Delete(key, false)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) CheckAndDelete(key string, oldValue string) error {
	_, err := c.client.CompareAndDelete(key, oldValue, 0)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdClient) CheckAndSet(key string, value string, ttl uint64, oldValue string) error {
	var err error
	if oldValue == "" {
		_, err = c.client.Create(key, value, ttl)
	} else {
		_, err = c.client.CompareAndSwap(key, value, ttl, oldValue, 0)
	}
	if err != nil {
		return err
	}
	return nil
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

func (c *etcdClient) watchWithoutRetry(key string, cancel chan bool, callBack func(string) error) error {
	var waitIndex uint64 = 1
	// First get the starting value of the key
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			err = callBack("")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		err = callBack(response.Node.Value)
		if err != nil {
			return err
		}
		waitIndex = response.Node.ModifiedIndex + 1
	}
	for {
		response, err := c.client.Watch(key, waitIndex, false, nil, cancel)
		if err != nil {
			if err == etcd.ErrWatchStoppedByUser {
				return ErrCancelled
			}
			return err
		}
		err = callBack(response.Node.Value)
		if err != nil {
			return err
		}
		waitIndex = response.Node.ModifiedIndex + 1
	}
}

func (c *etcdClient) watchAllWithoutRetry(key string, cancel chan bool, callBack func(map[string]string) error) error {
	var waitIndex uint64 = 1
	value := make(map[string]string)
	// First get the starting value of the key
	response, err := c.client.Get(key, false, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "100: Key not found") {
			err = callBack(nil)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		waitIndex = maxModifiedIndex(response.Node) + 1
		if nodeToMap(response.Node, value) {
			err = callBack(value)
			if err != nil {
				return err
			}
		}
	}
	for {
		response, err := c.client.Watch(key, waitIndex, true, nil, cancel)
		if err != nil {
			if err == etcd.ErrWatchStoppedByUser {
				return ErrCancelled
			}
			return err
		}
		responseModifiedIndex := maxModifiedIndex(response.Node)
		waitIndex = responseModifiedIndex + 1
		if nodeToMap(response.Node, value) {
			err = callBack(value)
			if err != nil {
				return err
			}
		}
	}
}
