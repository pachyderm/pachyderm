package address

type Addresser interface {
	GetMasterAddress(shard int) (string, error)
	GetSlaveAddresses(shard int) ([]string, error)
}

func NewSingleAddresser(address string) Addresser {
	return newSingleAddresser(address)
}

type ClientAddresser interface {
	GetServerAddress() (string, error)
}

func NewSingleClientAddresser(address string) ClientAddresser {
	return newSingleAddresser(address)
}
