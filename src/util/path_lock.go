package util

import (
	"fmt"
	"sync"
)

type PathLock interface {
	Lock(path string)
	Unlock(path string) error
}

func NewPathLock() PathLock {
	return newPathLock()
}

type pathLock struct {
	pathLocks map[string]*sync.Mutex
	lock      *sync.RWMutex
}

func newPathLock() *pathLock {
	return &pathLock{
		make(map[string]*sync.Mutex),
		&sync.RWMutex{},
	}
}

func (p *pathLock) Lock(path string) {
	p.lock.RLock()
	l := p.pathLocks[path]
	p.lock.RUnlock()
	if l == nil {
		p.lock.Lock()
		l = p.pathLocks[path]
		// check again in case someone got the write lock first
		if l == nil {
			l = &sync.Mutex{}
			p.pathLocks[path] = l
		}
		p.lock.Unlock()
	}
	l.Lock()
}

func (p *pathLock) Unlock(path string) error {
	p.lock.RLock()
	l := p.pathLocks[path]
	p.lock.RUnlock()
	if l == nil {
		return fmt.Errorf("no lock for path %s", path)
	}
	l.Unlock()
	return nil
}
