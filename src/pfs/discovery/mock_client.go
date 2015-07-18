package discovery

import (
	"fmt"
	"strings"
	"sync"
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
		return "", ErrNotFound
	}
	if record.directory {
		return "", ErrDirectory
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
	oldRecord, ok := c.records[key]
	if ok && oldRecord.directory {
		return ErrDirectory
	}
	c.records[key] = record{false, value}
	fmt.Println(c.records)
	return nil
}

func (c *mockClient) Create(key string, value string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.records[key]
	if ok {
		return ErrExists
	}
	c.records[key] = record{false, value}
	return nil
}

func (c *mockClient) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	oldRecord, ok := c.records[key]
	if !ok {
		return nil
	}
	if oldRecord.directory {
		return ErrDirectory
	}
	delete(c.records, key)
	return nil
}

func (c *mockClient) CheckAndSet(key string, value string, oldValue string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	oldRecord, ok := c.records[key]
	if !ok {
		return ErrNotFound
	}
	if oldRecord.directory {
		return ErrDirectory
	}
	if oldRecord.data != oldValue {
		return ErrPrecondition
	}
	c.records[key] = record{false, value}
	return nil
}
