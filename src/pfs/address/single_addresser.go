package address

type singleAddresser struct {
	address   string
	numShards int
}

func newSingleAddresser(address string, numShards int) *singleAddresser {
	return &singleAddresser{address, numShards}
}

func (l *singleAddresser) GetMasterShards(address string) (map[int]bool, error) {
	return l.newShardMap(), nil
}

func (l *singleAddresser) GetSlaveShards(address string) (map[int]bool, error) {
	return l.newShardMap(), nil
}

func (l *singleAddresser) GetAllAddresses() ([]string, error) {
	return []string{l.address}, nil
}

// want to make a new map each time since it is modifiable
func (l *singleAddresser) newShardMap() map[int]bool {
	m := make(map[int]bool, l.numShards)
	for i := 0; i < l.numShards; i++ {
		m[i] = true
	}
	return m
}
