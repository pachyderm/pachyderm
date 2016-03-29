package discovery

import (
	"strings"
	"time"
)

const (
	registryRefreshSec = 15
)

type registry struct {
	client    Client
	directory string
}

func newRegistry(client Client, directory string) *registry {
	return &registry{client, directory}
}

func (r *registry) Register(value string) <-chan error {
	errChan := make(chan error, 1)
	go r.register(value, errChan)
	return errChan
}

func (r *registry) GetAll() ([]string, error) {
	m, err := r.client.GetAll(r.directory)
	if err != nil {
		return nil, err
	}
	s := make([]string, len(m))
	i := 0
	for value := range m {
		s[i] = strings.TrimPrefix(value, r.directory+"/")
		i++
	}
	return s, nil
}

func (r *registry) register(value string, errChan chan<- error) {
	for {
		if err := r.client.Set(r.directory+"/"+value, value, registryRefreshSec*2); err != nil {
			errChan <- err
			return
		}
		time.Sleep(registryRefreshSec * time.Second)
	}
}
