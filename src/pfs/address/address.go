package address

type Addresser interface {
	GetMasterAddress(shard int) (string, error)
	GetSlaveAddresses(shard int) ([]string, error)
	GetAllAddresses() ([]string, error)
}

func NewSingleAddresser(address string) Addresser {
	return newSingleAddresser(address)
}
