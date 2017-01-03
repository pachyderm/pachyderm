package discovery

import (
	"context"
	"fmt"
)

// ErrCancelled is returned when an action is cancelled by the user
var ErrCancelled = fmt.Errorf("pachyderm: cancelled by user")

// Client defines Pachyderm's interface to key-value stores such as etcd.
type Client interface {
	// Close closes the underlying connection.
	Close() error
	// Get gets the value of a key
	// Keys can be directories of the form a/b/c, see etcd for details.
	// the error will be non-nil if the key does not exist.
	Get(ctx context.Context, key string) (string, error)
	// GetAll returns all of the keys in a directory and its subdirectories as
	// a map from absolute keys to values.
	// the map will be empty if no keys are found.
	GetAll(ctx context.Context, key string) (map[string]string, error)
	// Watch calls callBack with changes to a value
	Watch(ctx context.Context, key string, cancel chan bool, callBack func(string) error) error
	// WatchAll calls callBack with changes to a directory
	WatchAll(ctx context.Context, key string, cancel chan bool, callBack func(map[string]string) error) error
	// Set sets the value for a key.
	// ttl is in seconds.
	Set(ctx context.Context, key string, value string, ttl uint64) error
	// Delete deletes a key.
	Delete(ctx context.Context, key string) error
	// CheckAndSet is like Set but only succeeds if the key is already set to oldValue.
	// ttl is in seconds.
	CheckAndSet(ctx context.Context, key string, value string, ttl uint64, oldValue string) error
}

// NewEtcdClient creates an etcdClient with the given addresses.
func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}

// Registry is an object that allows a value to be registered as
// valid for the lifetime of a process, and allows all values
// registered to be retrieved.
type Registry interface {
	// Register registers the value. This will start a goroutine
	// that will constantly set the value as valid until the
	// process stops. If an error occurs, it will be returned on
	// the channel.
	Register(value string) <-chan error
	// GetAll gets all valid values.
	GetAll() ([]string, error)
}

// NewRegistry returns a new Registry using the given Client.
// All values will be placed inside the given directory.
func NewRegistry(client Client, directory string) Registry {
	return newRegistry(client, directory)
}
