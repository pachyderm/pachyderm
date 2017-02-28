// package watch implements better watch semantics on top of etcd.
// See this issue for the reasoning behind the package:
// https://github.com/coreos/etcd/issues/7362
package watch

import (
	"context"
	"sort"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/mirror"
	"github.com/gogo/protobuf/proto"
)

type EventType int

const (
	EventPut EventType = iota
	EventDelete
	EventError
)

type EventChan chan *Event

type Event struct {
	Key       []byte
	Value     []byte
	PrevKey   []byte
	PrevValue []byte
	Type      EventType
	Rev       int64
	Err       error
}

func (e *Event) Unmarshal(key *string, val proto.Message) error {
	*key = string(e.Key)
	return proto.UnmarshalText(string(e.Value), val)
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

func Watch(ctx context.Context, client *etcd.Client, prefix string) EventChan {
	eventCh := make(chan *Event)
	go func() {
		syncer := mirror.NewSyncer(client, prefix, 0)
		respCh, errCh := syncer.SyncBase(ctx)
	getBaseObjects:
		for {
			resp, ok := <-respCh
			if !ok {
				break getBaseObjects
			}
			// we sort the responses by modification time, in order to be
			// consistent with new events which by their nature are sorted
			// by modification time.
			sort.Sort(byModRev{resp})
			for _, kv := range resp.Kvs {
				eventCh <- &Event{
					Key:   kv.Key,
					Value: kv.Value,
					Type:  EventPut,
					Rev:   kv.ModRevision,
				}
			}
		}
		if err := <-errCh; err != nil {
			eventCh <- &Event{
				Type: EventError,
				Err:  err,
			}
			close(eventCh)
			return
		}
		watchCh := syncer.SyncUpdates(ctx)
		for {
			resp := <-watchCh
			if err := resp.Err(); err != nil {
				eventCh <- &Event{
					Type: EventError,
					Err:  err,
				}
				close(eventCh)
				return
			}
			for _, etcdEv := range resp.Events {
				ev := &Event{
					Key:   etcdEv.Kv.Key,
					Value: etcdEv.Kv.Value,
					Rev:   etcdEv.Kv.ModRevision,
				}
				if etcdEv.PrevKv != nil {
					ev.PrevKey = etcdEv.PrevKv.Key
					ev.PrevValue = etcdEv.PrevKv.Value
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
	}()
	return eventCh
}
