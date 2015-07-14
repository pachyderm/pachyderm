package address

type Addresser interface {
	GetMasterAddress(shard int) (string, error)
	GetSlaveAddresses(shard int) ([]string, error)
}

func NewLocalAddresser(port uint16) Addresser {
	return newLocalAddresser(port)
}
