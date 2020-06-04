package plugin

import (
	"os"
	stdplugin "plugin"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

var (
	loadedPlugins map[*pfs.File]*stdplugin.Plugin = map[*pfs.File]*stdplugin.Plugin{} // TODO(ys): will lookups here work fine, given pfs.File is behind a pointer?
	mu sync.Mutex
)

func Load(f *pfs.File, getter func(*pfs.File) (string, error)) (*stdplugin.Plugin, error) {
	mu.Lock()
	p, ok := loadedPlugins[f]
	if ok {
		mu.Unlock()
		return p, nil
	}
	mu.Unlock()

	filepath, err := getter(f)
	if err != nil {
		return nil, err
	}
	defer os.Remove(filepath)

	mu.Lock()
	defer mu.Unlock()

	// double-check that the plugin didn't load in another goroutine since we
	// last locked and checked
	p, ok = loadedPlugins[f]
	if ok {
		return p, nil
	}

	p, err = stdplugin.Open(filepath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open delimiter plugin file")
	}

	loadedPlugins[f] = p
	return p, nil
}