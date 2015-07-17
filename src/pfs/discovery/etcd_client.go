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
