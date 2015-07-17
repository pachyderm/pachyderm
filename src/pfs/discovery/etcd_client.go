package discovery

type etcdClient struct {
	addresses []string
}

func newEtcdClient(addresses ...string) *etcdClient {
	return &etcdClient{addresses}
}

func (c *etcdClient) Close() error {
	return nil
}

func (c *etcdClient) Get(key string) (string, error) {
	return "", nil
}

func (c *etcdClient) GetAll(key string) (map[string]string, error) {
	return nil, nil
}

func (c *etcdClient) Set(key string, value string) error {
	return nil
}

func (c *etcdClient) Create(key string, value string) error {
	return nil
}

func (c *etcdClient) Delete(key string) error {
	return nil
}

func (c *etcdClient) CheckAndSet(key string, value string, oldValue string) error {
	return nil
}
