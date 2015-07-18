package route

import (
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs/discovery"
)

type discoveryAddresser struct {
	discoveryClient discovery.Client
}

func newDiscoveryAddresser(discoveryClient discovery.Client) *discoveryAddresser {
	return &discoveryAddresser{discoveryClient}
}

func (l *discoveryAddresser) GetMasterShards(address string) (map[int]bool, error) {
	value, err := l.discoveryClient.Get(address + "-master")
	if err != nil {
		return nil, err
	}
	return l.newShardMap(value)
}

func (l *discoveryAddresser) GetSlaveShards(address string) (map[int]bool, error) {
	//value, err := l.discoveryClient.Get(address + "-slave")
	//if err != nil {
	//return nil, err
	//}
	//return l.newShardMap(value)
	return make(map[int]bool), nil
}

func (l *discoveryAddresser) GetAllAddresses() ([]string, error) {
	value, err := l.discoveryClient.Get("all-addresses")
	if err != nil {
		return nil, err
	}
	return strings.Split(value, ","), nil
}

func (l *discoveryAddresser) newShardMap(value string) (map[int]bool, error) {
	split := strings.Split(value, ",")
	m := make(map[int]bool, len(split))
	for _, s := range split {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		m[int(i)] = true
	}
	return m, nil
}
