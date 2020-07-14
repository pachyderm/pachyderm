package types

import (
	"reflect"
	"sync"
)

// Registry for Any
func NewTypesRegistry() *TypesRegistry {
	return &TypesRegistry{
		registry: make(map[string]reflect.Type),
	}
}

type TypesRegistry struct {
	registry map[string]reflect.Type
	sync.RWMutex
}

func (this *TypesRegistry) Get(name string) (reflect.Type, bool) {
	this.RLock()
	defer this.RUnlock()
	t, ok := this.registry[name]
	return t, ok
}

func (this *TypesRegistry) Register(name string, t reflect.Type) {
	this.Lock()
	this.registry[name] = t
	this.Unlock()
}

// Registry for Perms bitmaps
func NewPermsRegistry() *PermsRegistry {
	return &PermsRegistry{
		registry: make(map[string]*PermissionsBitmap),
	}
}

type PermsRegistry struct {
	registry map[string]*PermissionsBitmap
	sync.RWMutex
}

func (p *PermsRegistry) Get(name string) (*PermissionsBitmap, bool) {
	p.RLock()
	defer p.RUnlock()
	r, ok := p.registry[name]
	return r, ok
}

func (p *PermsRegistry) Register(name string, bmap *PermissionsBitmap) {
	p.Lock()
	p.registry[name] = bmap
	p.Unlock()
}
