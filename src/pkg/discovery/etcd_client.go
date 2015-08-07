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
		return "", false, err
	}
	return response.Node.Value, true, nil
}

func (c *etcdClient) GetAll(key string) (map[string]string, error) {
	response, err := c.client.Get(key, false, true)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, node := range response.Node.Nodes {
		result[strings.TrimPrefix(node.Key, "/")] = node.Value
	}
	return result, nil
}

func (c *etcdClient) Set(key string, value string, ttl uint64) error {
	_, err := c.client.Set(key, value, ttl)
	return err
}

func (c *etcdClient) Create(key string, value string, ttl uint64) error {
	_, err := c.client.Create(key, value, ttl)
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
