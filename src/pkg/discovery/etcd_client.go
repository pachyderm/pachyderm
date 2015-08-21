package discovery

import (
	"strings"
	"sync"

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

func (c *etcdClient) Watch(key string, cancel chan bool, callBack func(string) error) (retErr error) {
	localCancel := make(chan bool)
	var once sync.Once
	var errOnce sync.Once
	go func() {
		select {
		case <-cancel:
			once.Do(func() { close(localCancel) })
		case <-localCancel:
		}
	}()
	receiver := make(chan *etcd.Response)
	defer close(receiver)
	go func() {
		for response := range receiver {
			if err := callBack(response.Node.Value); err != nil {
				errOnce.Do(func() { retErr = err })
				once.Do(func() { close(localCancel) })
			}
		}
	}()
	if _, err := c.client.Watch(key, 0, false, receiver, localCancel); err != nil {
		errOnce.Do(func() { retErr = err })
	}
	return
}

func (c *etcdClient) WatchAll(key string, cancel chan bool, callBack func(map[string]string) error) (retErr error) {
	localCancel := make(chan bool)
	var errOnce sync.Once
	var once sync.Once
	go func() {
		select {
		case <-cancel:
			once.Do(func() { close(localCancel) })
		case <-localCancel:
		}
	}()
	receiver := make(chan *etcd.Response)
	defer close(receiver)
	go func() {
		for response := range receiver {
			value := make(map[string]string)
			nodeToMap(response.Node, value)
			if err := callBack(value); err != nil && retErr == nil {
				errOnce.Do(func() { retErr = err })
				once.Do(func() { close(localCancel) })
			}
		}
	}()
	if _, err := c.client.Watch(key, 0, true, receiver, cancel); err != nil {
		errOnce.Do(func() { retErr = err })
	}
	return
}

func (c *etcdClient) Set(key string, value string, ttl uint64) error {
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

func nodeToMap(node *etcd.Node, out map[string]string) {
	if !node.Dir {
		out[strings.TrimPrefix(node.Key, "/")] = node.Value
	} else {
		for _, node := range node.Nodes {
			nodeToMap(node, out)
		}
	}
}
