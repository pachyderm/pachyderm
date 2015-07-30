package discovery

type Client interface {
	// Close closes the underlying connection
	Close() error
	// Get gets the value of a key
	Get(key string) (string, error)
	// GetAll returns all of the keys in a directory and its subdirectories as
	// a map from absolute keys to values.
	GetAll(key string) (map[string]string, error)
	// Set sets the value for a key
	Set(key string, value string, ttl uint64) error
	// Delete deletes a key.
	Delete(key string) error
	// Create is like Set but only succeeds if the key doesn't already exist
	Create(key string, value string, ttl uint64) error
	// CheckAndSet is like Set but only succeeds if the key is already set to oldValue
	CheckAndSet(key string, value string, ttl uint64, oldValue string) error
}

func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}

func NewMockClient() Client {
	return newMockClient()
}
