// package watch implements better watch semantics on top of etcd.
// See this issue for the reasoning behind the package:
// https://github.com/coreos/etcd/issues/7362
package watch

import (
	"context"
	"sort"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/mirror"
)

type EventType int

const (
	EventPut EventType = iota
	EventDelete
)

type EventChan chan *Event

type Event struct {
	Key   string
	Value string
	Type  EventType
	Rev   int64
}

// sort the key-value pairs by revision time
type byModRev struct{ etcd.GetResponse }

func (s byModRev) Len() int {
	return len(s.GetResponse.Kvs)
}

func (s byModRev) Swap(i, j int) {
	s.GetResponse.Kvs[i], s.GetResponse.Kvs[j] = s.GetResponse.Kvs[j], s.GetResponse.Kvs[i]
}

func (s byModRev) Less(i, j int) bool {
	return s.GetResponse.Kvs[i].ModRevision < s.GetResponse.Kvs[j].ModRevision
}

func Watch(ctx context.Context, client *etcd.Client, prefix string) (EventChan, chan error) {
	eventCh := make(chan Event)
	errCh := make(chan error)
	go func() (retErr error) {
		defer func() {
			if retErr != nil {
				errCh <- retErr
			}
			close(errCh)
			close(eventCh)
		}()
		syncer := mirror.NewSyncer(client, prefix, 0)
		respCh, errCh := syncer.SyncBase(ctx)
	getBaseObjects:
		for {
			select {
			case resp, ok := <-respCh:
				if !ok {
					break getBaseObjects
				}
				// we sort the responses by modification time, in order to be
				// consistent with new events which by their nature are sorted
				// by modification time.
				sort.Sort(byModRev{resp})
				for _, kv := range resp.Kvs {
					eventCh <- &Event{
						key:   string(kv.Key),
						value: string(kv.Value),
						typ:   EventPut,
						rev:   kv.ModRevision,
					}
				}
			case err := <-errCh:
				if err != nil {
					return err
				}
			}
		}
		watchCh := syncer.SyncUpdates(ctx)
		for {
			resp := <-watchCh
			if err := resp.Err(); err != nil {
				return err
			}
			for _, etcdEv := range resp.Events {
				ev := &Event{
					Key:   string(etcdEv.Kv.Key),
					Value: string(etcdEv.Kv.Value),
					Rev:   etcdEv.Kv.ModRevision,
				}
				switch etcdEv.Type {
				case etcd.EventTypePut:
					ev.Type = EventPut
				case etcd.EventTypeDelete:
					ev.Type = EventDelete
				}
				eventCh <- ev
			}
		}
		return nil
	}()
	return eventCh, errCh
}
