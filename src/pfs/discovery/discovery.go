package discovery

import (
	"errors"
)

var (
	ErrNetwork       = errors.New("pachyderm: network failure")
	ErrNotFound      = errors.New("pachyderm: Not found")
	ErrIsDir         = errors.New("pachyderm: Found dir")
	ErrWrongOldValue = errors.New("pachyderm: Wrong old value")
	ErrExists        = errors.New("pachyderm: Key already exists")
)

type Client interface {
	// Close closes the underlying connection
	// Errors: ErrNetwork
	Close() error
	// Get gets the value of a key
	// Errors: ErrNetwork, ErrNotFound, ErrIsDir
	Get(key string) (string, error)
	// GetAll returns all of the keys in a directory and its subdirectories as
	// a map from absolute keys to values.
	// Errors: ErrNetwork, ErrNotFound
	GetAll(key string) (map[string]string, error)
	// Set sets the value for a key
	// Errors: ErrNetwork, ErrIsDir
	Set(key string, value string) error
	// Delete deletes a key.
	// Errors: ErrNetwork, ErrNotFound, ErrIsDir
	Delete(key string) error
	// Create is like Set but only succeeds if the key doesn't already exist
	// Errors: ErrNetwork, ErrExists
	Create(key string, value string) error
	// CheckAndSet is like Set but only succeeds if the key is already set to oldValue
	// Errors: ErrNetwork, ErrNotFound, ErrIsDir, ErrWrongOldValue
	CheckAndSet(key string, value string, oldValue string) error
}

func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}

func NewMockClient() *mockClient {
	return newMockClient()
}
