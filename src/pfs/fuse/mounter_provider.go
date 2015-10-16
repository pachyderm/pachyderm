package fuse

import (
	"sync"
	"sync/atomic"

	"go.pachyderm.com/pachyderm/src/pfs"
	"google.golang.org/grpc"
)

type mounterProvider struct {
	pfsAddress string
	once       *sync.Once
	value      *atomic.Value
	err        *atomic.Value
}

func newMounterProvider(pfsAddress string) *mounterProvider {
	return &mounterProvider{
		pfsAddress,
		&sync.Once{},
		&atomic.Value{},
		&atomic.Value{},
	}
}

func (m *mounterProvider) Get() (Mounter, error) {
	m.once.Do(func() {
		value, err := m.getOnce()
		if value != nil {
			m.value.Store(value)
		}
		if err != nil {
			m.err.Store(err)
		}
	})
	mounterObj := m.value.Load()
	errObj := m.err.Load()
	if errObj != nil {
		return nil, errObj.(error)
	}
	return mounterObj.(Mounter), nil
}

func (m *mounterProvider) getOnce() (Mounter, error) {
	clientConn, err := grpc.Dial(m.pfsAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return NewMounter(
		pfs.NewAPIClient(
			clientConn,
		),
	), nil
}
