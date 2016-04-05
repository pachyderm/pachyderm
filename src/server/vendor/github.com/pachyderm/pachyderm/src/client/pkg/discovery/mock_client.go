package discovery

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
)

type record struct {
	directory bool
	data      string
	expires   time.Time
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

func (c *mockClient) Get(key string) (string, bool, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	now := time.Now()
	record, ok := c.records[key]
	if !ok {
		return "", false, nil
	}
	if record.directory {
		return "", false, fmt.Errorf("pachyderm: key %s is directory", key)
	}
	if (record.expires != time.Time{}) && now.After(record.expires) {
		delete(c.records, key)
		return "", false, nil
	}
	return record.data, true, nil
}

func (c *mockClient) GetAll(keyPrefix string) (map[string]string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	now := time.Now()
	result := make(map[string]string)
	for key, record := range c.records {
		if (record.expires != time.Time{}) && now.After(record.expires) {
			delete(c.records, key)
		}
		if strings.HasPrefix(key, keyPrefix) && !record.directory {
			result[key] = record.data
		}
	}
	return result, nil
}

func (c *mockClient) Watch(key string, cancel chan bool, callBack func(string) error) error {
	// TODO jdoliner
	return nil
}

func (c *mockClient) WatchAll(key string, cancel chan bool, callBack func(map[string]string) error) error {
	// TODO jdoliner
	return nil
}

func (c *mockClient) Set(key string, value string, ttl uint64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.unsafeSet(key, value, ttl)
}

func (c *mockClient) Create(key string, value string, ttl uint64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.records[key]
	if ok {
		return fmt.Errorf("pachyderm: key %s already exists", key)
	}
	return c.unsafeSet(key, value, ttl)
}

func (c *mockClient) CreateInDir(dir string, value string, ttl uint64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	key := path.Join(dir, uuid.NewWithoutDashes())
	return c.unsafeSet(key, value, ttl)
}

func (c *mockClient) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	oldRecord, ok := c.records[key]
	if !ok {
		return nil
	}
	if oldRecord.directory {
		return fmt.Errorf("pachyderm: can't delete directory %s", key)
	}
	delete(c.records, key)
	return nil
}

func (c *mockClient) CheckAndSet(key string, value string, ttl uint64, oldValue string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	oldRecord, ok := c.records[key]
	if !ok {
		return fmt.Errorf("pachyderm: key %s not found", key)
	}
	if oldRecord.directory {
		return fmt.Errorf("pachyderm: can't set directory %s", key)
	}
	if oldRecord.data != oldValue {
		return fmt.Errorf("pachyderm: precondition not met for %s", key)
	}
	return c.unsafeSet(key, value, ttl)
}

func (c *mockClient) unsafeSet(key string, value string, ttl uint64) error {
	var expires time.Time
	if ttl != 0 {
		expires = time.Now().Add(time.Second * time.Duration(ttl))
	}
	parts := strings.Split(key, "/")
	for i := range parts[:len(parts)-1] {
		oldRecord, ok := c.records["/"+path.Join(parts[:i]...)]
		if ok && !oldRecord.directory {
			return fmt.Errorf("pachyderm: key %s not found", key)
		}
		if !ok {
			c.records["/"+path.Join(parts[:i]...)] = record{true, "", expires}
		}
	}
	oldRecord, ok := c.records[key]
	if ok && oldRecord.directory {
		return fmt.Errorf("pachyderm: can't set directory %s", key)
	}
	c.records[key] = record{false, value, expires}
	return nil
}

func (c *mockClient) Hold(key string, value string, oldValue string, cancel chan bool) error {
	// TODO jdoliner
	return nil
}
