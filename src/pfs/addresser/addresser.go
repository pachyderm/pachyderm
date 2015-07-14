package addresser

type Addresser interface {
	GetAddress(shard int) (string, error)
}

func NewLocalAddresser(port uint16) Addresser {
	return nil
}
