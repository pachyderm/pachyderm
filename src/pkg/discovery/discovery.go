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
	Watch(key string, stop chan bool, callBack func(string)) error
	// WatchAll calls callBack with changes to a directory
	WatchAll(key string, stop chan bool, callBack func(map[string]string)) error
	// Set sets the value for a key.
	// ttl is in seconds.
	Set(key string, value string, ttl uint64) error
	// Delete deletes a key.
	Delete(key string) error
	// Create is like Set but only succeeds if the key doesn't already exist.
	// ttl is in seconds.
	Create(key string, value string, ttl uint64) error
	// CreateInDir is like Set but it generates a key inside dir.
	CreateInDir(dir string, value string, ttl uint64) error
	// CheckAndSet is like Set but only succeeds if the key is already set to oldValue.
	// ttl is in seconds.
	CheckAndSet(key string, value string, ttl uint64, oldValue string) error
}

func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}

func NewMockClient() Client {
	return newMockClient()
}
