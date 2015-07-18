package discovery

import (
	"path"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type record struct {
	directory bool
	data      string
}

type mockClient struct {
	records map[string]record
	lock    sync.RWMutex
}

func newMockClient() *mockClient {
	return &mockClient{
		make(map[string]record),
		sync.RWMutex{},
	}
}

func (c *mockClient) Close() error {
	return nil
}

func (c *mockClient) Get(key string) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	record, ok := c.records[key]
	if !ok {
		return "", pfs.ErrDiscoveryNotFound
	}
	if record.directory {
		return "", pfs.ErrDiscoveryNotValue
	}
	return record.data, nil
}

func (c *mockClient) GetAll(key string) (map[string]string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	result := make(map[string]string)
	for k, v := range c.records {
		if strings.HasPrefix(k, key) && !v.directory {
			result[k] = v.data
		}
	}
	return result, nil
}

func (c *mockClient) Set(key string, value string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.unsafeSet(key, value)
}

func (c *mockClient) Create(key string, value string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.records[key]
	if ok {
		return pfs.ErrDiscoveryKeyAlreadyExists
	}
	return c.unsafeSet(key, value)
}

func (c *mockClient) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	oldRecord, ok := c.records[key]
	if !ok {
		return nil
	}
	if oldRecord.directory {
		return pfs.ErrDiscoveryNotValue
	}
	delete(c.records, key)
	return nil
}

func (c *mockClient) CheckAndSet(key string, value string, oldValue string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	oldRecord, ok := c.records[key]
	if !ok {
		return pfs.ErrDiscoveryNotFound
	}
	if oldRecord.directory {
		return pfs.ErrDiscoveryNotValue
	}
	if oldRecord.data != oldValue {
		return pfs.ErrDiscoveryPreconditionNotMet
	}
	return c.unsafeSet(key, value)
}

func (c *mockClient) unsafeSet(key string, value string) error {
	parts := strings.Split(key, "/")
	for i, _ := range parts[:len(parts)-1] {
		oldRecord, ok := c.records["/"+path.Join(parts[:i]...)]
		if ok && oldRecord.directory {
			return pfs.ErrDiscoveryNotDirectory
		}
		if !ok {
			c.records["/"+path.Join(parts[:i]...)] = record{true, ""}
		}
	}
	oldRecord, ok := c.records[key]
	if ok && oldRecord.directory {
		return pfs.ErrDiscoveryNotValue
	}
	c.records[key] = record{false, value}
	return nil
}
