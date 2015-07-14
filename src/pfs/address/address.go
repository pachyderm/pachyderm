package address

type Addresser interface {
	GetMasterAddress(shard int) (string, error)
	GetSlaveAddresses(shard int) ([]string, error)
	SetAddress(shard int, address string, master bool) error
}

func NewLocalAddresser(port uint16) Addresser {
	return nil
}
