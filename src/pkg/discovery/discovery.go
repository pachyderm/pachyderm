package discovery

type Client interface {
	// Close closes the underlying connection.
	Close() error
	// Get gets the value of a key
	// Keys can be directories of the form a/b/c, see etcd for details.
	// the bool will be false if the key does not exist.
	Get(key string) (string, bool, error)
	// GetAll returns all of the keys in a directory and its subdirectories as
	// a map from absolute keys to values.
	// the map will be empty if no keys are found.
	GetAll(key string) (map[string]string, error)
	// Watch calls callBack with changes to a value
	Watch(key string, cancel chan bool, callBack func(string) (uint64, error)) error
	// WatchAll calls callBack with changes to a directory
	WatchAll(key string, cancel chan bool, callBack func(map[string]string) (uint64, error)) error
	// Set sets the value for a key.
	// ttl is in seconds.
	Set(key string, value string, ttl uint64) (uint64, error)
	// Delete deletes a key.
	Delete(key string) (uint64, error)
	// CheckAndDelete deletes a key only if its value matches oldValue
	CheckAndDelete(key string, oldValue string) (uint64, error)
	// Create is like Set but only succeeds if the key doesn't already exist.
	// ttl is in seconds.
	Create(key string, value string, ttl uint64) (uint64, error)
	// CreateInDir is like Set but it generates a key inside dir.
	CreateInDir(dir string, value string, ttl uint64) (uint64, error)
	// CheckAndSet is like Set but only succeeds if the key is already set to oldValue.
	// ttl is in seconds.
	CheckAndSet(key string, value string, ttl uint64, oldValue string) (uint64, error)
	// Hold periodically updates a key with a nonzero ttl value.
	// This is useful for advertising a service since it will automatically go
	// away if the service dies.
	// Hold assumes that key is ALREADY set to value.
	Hold(key string, value string, ttl uint64, cancel chan bool) error
}

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
