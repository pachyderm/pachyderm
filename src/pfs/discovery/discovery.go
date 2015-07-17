package discovery

type Client interface {
	Close() error
	Get(key string) (string, error)
	GetAll(key string) ([]string, error)
	Set(key string, value string) error
	Create(key string, value string) error
	Delete(key string) error
	CheckAndSet(key string, value string, oldValue string) error
}

func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}
