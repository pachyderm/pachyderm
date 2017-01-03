package discovery

import (
	"context"
	"errors"
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
	return &etcdClient{client}, nil
}

func (c *etcdClient) Close() error {
	c.client.Close()
	return nil
}

func (c *etcdClient) Get(ctx context.Context, key string) (string, error) {
	response, err := c.client.Get(ctx, key)
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
		return nil, err
	}
	if response.Count < 1 {
		return nil, KeyNotFoundErr
	}
	result := make(map[string]string, 0)
	for _, kv := range response.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}
	return result, nil
}

func (c *etcdClient) Watch(ctx context.Context, key string, cancel chan bool, callback func(string) error) error {
	rch := c.client.Watch(ctx, key)
	for rsp := range rch {
		if err := rsp.Err(); err != nil {
			return err
		}
		for _, ev := range rsp.Events {
			if err := callback(string(ev.Kv.Value)); err != nil {
				return err
			}
		}
	}
	return errors.New("unreachable")
}

func (c *etcdClient) WatchAll(ctx context.Context, key string, cancel chan bool, callback func(map[string]string) error) error {
	rch := c.client.Watch(ctx, key)
	for rsp := range rch {
		if err := rsp.Err(); err != nil {
			return err
		}
		result := make(map[string]string, 0)
		for _, ev := range rsp.Events {
			result[string(ev.Kv.Key)] = string(ev.Kv.Value)
		}
		if err := callback(result); err != nil {
			return err
		}
	}
	return errors.New("unreachable")
}

func (c *etcdClient) Set(ctx context.Context, key string, value string, ttl uint64) error {
	lease, err := c.client.Grant(ctx, int64(ttl))
	if err != nil {
		return err
	}
	_, err = c.client.Put(ctx, key, value, etcd.WithLease(lease.ID))
	return err
}

func (c *etcdClient) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	return err
}

func (c *etcdClient) CheckAndSet(ctx context.Context, key string, value string, ttl uint64, oldValue string) error {
	lease, err := c.client.Grant(ctx, int64(ttl))
	if err != nil {
		return err
	}
	_, err = c.client.Txn(ctx).If(etcd.Compare(etcd.Value(key), "=", oldValue)).Then(etcd.OpPut(key, value, etcd.WithLease(lease.ID))).Commit()
	return err
}
