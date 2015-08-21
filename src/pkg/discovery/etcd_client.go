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

func (c *etcdClient) Watch(key string, stop chan bool, callBack func(string) error) error {
	receiver := make(chan *etcd.Response)
	defer close(receiver)
	go func() {
		for response := range receiver {
			callBack(response.Node.Value)
		}
	}()
	_, err := c.client.Watch(key, 0, false, receiver, stop)
	return err
}

func (c *etcdClient) WatchAll(key string, stop chan bool, callBack func(map[string]string) error) error {
	receiver := make(chan *etcd.Response)
	defer close(receiver)
	go func() {
		for response := range receiver {
			value := make(map[string]string)
			nodeToMap(response.Node, value)
			callBack(value)
		}
	}()
	_, err := c.client.Watch(key, 0, true, receiver, stop)
	return err
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
