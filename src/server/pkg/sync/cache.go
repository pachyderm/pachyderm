package sync

import (
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	lru "github.com/hashicorp/golang-lru"
	pachclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type Cache struct {
	*lru.Cache
	storageRoot string
	puller      *Puller
}

func NewCache(size int, storageRoot string) (*Cache, error) {
	c, err := lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		if err := os.RemoveAll(key.(string)); err != nil {
			logrus.Errorf("failed to remove %s", filepath.Join(storageRoot, key.(string)))
		}
	})
	if err != nil {
		return nil, err
	}
	return &Cache{
		Cache:       c,
		storageRoot: storageRoot,
		puller:      NewPuller(),
	}, nil
}

func (c *Cache) key(file *pfs.File) string {
	return filepath.Join(c.storageRoot, file.Commit.Repo.Name, file.Commit.ID, file.Path)
}

func (c *Cache) PullFile(pachClient *pachclient.APIClient, file *pfs.File, concurrency int) (string, error) {
	key := c.key(file)
	if c.Contains(key) {
		return key, nil
	}
	if err := c.puller.Pull(pachClient, key, file.Commit.Repo.Name, file.Commit.ID, file.Path, false, false, concurrency, nil, ""); err != nil {
		return "", err
	}
	return key, nil
}
