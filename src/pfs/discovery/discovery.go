package discovery

type Client interface {
	Close() error
}

func NewEtcdClient(addresses ...string) Client {
	return newEtcdClient(addresses...)
}
