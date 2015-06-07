package concurrency

import (
	"sync"
)

type PathLock struct {
	pathLocks map[string]*sync.Mutex
	lock      sync.RWMutex
}

func NewPathLock() *PathLock {
	return &PathLock{pathLocks: make(map[string]*sync.Mutex)}
}

func (p *PathLock) Lock(path string) {
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

func (p *PathLock) Unlock(path string) {
	p.lock.RLock()
	l := p.pathLocks[path]
	p.lock.RUnlock()
	if l == nil {
		panic("Trying to unlock unknown lock.")
	}
	l.Unlock()
}
