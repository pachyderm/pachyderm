package discovery

import (
	"context"
	"errors"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
)

var (
	KeyNotFoundErr = errors.New("key not found")
)

type etcdClient struct {
	client *etcd.Client
}

func newEtcdClient(addresses ...string) (*etcdClient, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   addresses,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &etcdClient{client}
}

func (c *etcdClient) Close() error {
	c.client.Close()
	return nil
}

func (c *etcdClient) Get(ctx context.Context, key string) (string, error) {
	response, err := c.client.Get(key)
	if err != nil {
		return "", err
	}
	if response.Count < 1 {
		return "", KeyNotFoundErr
	}
	return string(response.Kvs[0].Value), nil
}

func (c *etcdClient) GetAll(ctx context.Context, key string) (map[string]string, error) {
	response, err := c.client.Get(ctx, key, etcd.WithPrefix())
	if err != nil {
		return "", err
	}
	if response.Count < 1 {
		return "", KeyNotFoundErr
	}
	result := make(map[string]string, 0)
	for _, kv := range response.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}
	return result, nil
}

func (c *etcdClient) Watch(ctx context.Context, key string, cancel chan bool, callBack func(string) error) error {
	rch := c.client.Watch(ctx, key)
	for rsp := range rch {
		if err := rsp.Err(); err != nil {
			return err
		}
		for _, ev := range rsp.Events {
			if err != callback(string(ev.Kv.Value)); err != nil {
				return err
			}
		}
	}
	return errors.New("unreachable")
}

func (c *etcdClient) WatchAll(ctx context.Context, key string, cancel chan bool, callBack func(map[string]string) error) error {
	rch := c.client.Watch(ctx, key)
	for rsp := range rch {
		if err := rsp.Err(); err != nil {
			return err
		}
		result := make(map[string]string, 0)
		for _, ev := range rsp.Events {
			result[string(ev.Kv.Key)] = string(ev.Kv.Value)
		}
		if err != callback(result); err != nil {
			return err
		}
	}
	return errors.New("unreachable")
}

func (c *etcdClient) Set(ctx context.Context, key string, value string, ttl uint64) error {
	lease, err := c.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	_, err := c.client.Put(ctx, key, value, etcd.WithLease(lease.ID))
	return err
}

func (c *etcdClient) Delete(key string) error {
	_, err := c.client.Delete(key, false)
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
