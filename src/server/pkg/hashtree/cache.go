package hashtree

import (
	"reflect"

	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

// Cache is an LRU cache for hashtrees.
type Cache struct {
	*lru.Cache
}

// NewCache creates a new cache.
func NewCache(size int) (*Cache, error) {
	c, err := lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		tree, ok := value.(*dbHashTree)
		if !ok {
			logrus.Infof("non hashtree slice value of type: %v", reflect.TypeOf(value))
			return
		}
		if err := tree.Destroy(); err != nil {
			logrus.Infof("failed to destroy hashtree: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}
	return &Cache{c}, nil
}
