package discovery

import (
	"errors"
)

var (
	ErrNetwork      = errors.New("pachyderm: network failure")
	ErrNotFound     = errors.New("pachyderm: Not found")
	ErrDirectory    = errors.New("pachyderm: Found directory instead of value")
	ErrExists       = errors.New("pachyderm: Key already exists")
	ErrPrecondition = errors.New("pachyderm: Precondition not met")
)

type Client interface {
	// Close closes the underlying connection
	// Errors: ErrNetwork
	Close() error
	// Get gets the value of a key
	// Errors: ErrNetwork, ErrNotFound, ErrDirectory
	Get(key string) (string, error)
	// GetAll returns all of the keys in a directory and its subdirectories as
	// a map from absolute keys to values.
	// Errors: ErrNetwork
	GetAll(key string) (map[string]string, error)
	// Set sets the value for a key
	// Errors: ErrNetwork, ErrDirectory
	Set(key string, value string) error
	// Delete deletes a key.
	// Errors: ErrNetwork, ErrNotFound, ErrDirectory
	Delete(key string) error
	// Create is like Set but only succeeds if the key doesn't already exist
	// Errors: ErrNetwork, ErrExists
	Create(key string, value string) error
	// CheckAndSet is like Set but only succeeds if the key is already set to oldValue
	// Errors: ErrNetwork, ErrNotFound, ErrDirectory, ErrPrecondition
	CheckAndSet(key string, value string, oldValue string) error
}

func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}

func NewMockClient() Client {
	return newMockClient()
}
