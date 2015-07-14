package address

import "fmt"

type localAddresser struct {
	port uint16
}

func newLocalAddresser(port uint16) *localAddresser {
	return &localAddresser{port}
}

func (l *localAddresser) GetMasterAddress(shard int) (string, error) {
	return fmt.Sprintf("0.0.0.0:%d", l.port), nil
}

func (l *localAddresser) GetSlaveAddresses(shard int) ([]string, error) {
	return []string{fmt.Sprintf("0.0.0.0:%d", l.port)}, nil
}
