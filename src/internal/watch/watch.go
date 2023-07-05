// Package watch implements better watch semantics on top of etcd.
// See this issue for the reasoning behind the package:
// https://github.com/coreos/etcd/issues/7362
package watch

import (
	"bytes"
	"context"
	"reflect"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
)

// EventType is the type of event
type EventType int

const (
	// EventPut happens when an item is added
	EventPut EventType = iota
	// EventDelete happens when an item is removed
	EventDelete
	// EventError happens when an error occurred
	EventError
)

// Event is an event that occurred to an item in etcd.
type Event struct {
	Key      []byte
	Value    []byte
	Type     EventType
	Rev      int64
	Ver      int64
	Err      error
	Template proto.Message
}

// Unmarshal unmarshals the item in an event into a protobuf message.
func (e *Event) Unmarshal(key *string, val proto.Message) error {
	if e.Value == nil {
		return errors.Errorf("Cannot unmarshal an event with a null value, type: %v", e.Type)
	}
	if err := CheckType(e.Template, val); err != nil {
		return err
	}
	*key = string(e.Key)
	return errors.EnsureStack(proto.Unmarshal(e.Value, val))
}

// Watcher ...
type Watcher interface {
	// Watch returns a channel that delivers events
	Watch() <-chan *Event
	// Close this channel when you are done receiving events
	Close()
}

type etcdWatcher struct {
	eventCh chan *Event
	done    chan struct{}
}

func (w *etcdWatcher) Watch() <-chan *Event {
	return w.eventCh
}

func (w *etcdWatcher) Close() {
	close(w.done)
}

// NewEtcdWatcher watches a given etcd prefix for events.
func NewEtcdWatcher(ctx context.Context, client *etcd.Client, trimPrefix, prefix string, template proto.Message, opts ...Option) (Watcher, error) {
	eventCh := make(chan *Event)
	done := make(chan struct{})
	options := SumOptions(opts...)

	// First list the collection to get the current items
	getOptions := []etcd.OpOption{etcd.WithPrefix(), etcd.WithSort(options.SortTarget, options.SortOrder)}
	resp, err := client.Get(ctx, prefix, getOptions...)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	nextRevision := resp.Header.Revision + 1
	watchOptions := func(rev int64) []etcd.OpOption {
		result := []etcd.OpOption{etcd.WithPrefix(), etcd.WithRev(rev)}
		if !options.IncludePut {
			result = append(result, etcd.WithFilterPut())
		}
		if !options.IncludeDelete {
			result = append(result, etcd.WithFilterDelete())
		}
		return result
	}

	internalWatcher := etcd.NewWatcher(client)
	// Issue a watch that uses the revision timestamp returned by the
	// Get request earlier.  That way even if some items are added between
	// when we list the collection and when we start watching the collection,
	// we won't miss any items.

	rch := internalWatcher.Watch(ctx, prefix, watchOptions(nextRevision)...)

	go func() (retErr error) {
		defer func() {
			if retErr != nil {
				select {
				case eventCh <- &Event{
					Err:  retErr,
					Type: EventError,
				}:
				case <-done:
				}
			}
			close(eventCh)
			internalWatcher.Close()
		}()
		for _, etcdKv := range resp.Kvs {
			e := &Event{
				Key:      bytes.TrimPrefix(etcdKv.Key, []byte(trimPrefix)),
				Value:    etcdKv.Value,
				Type:     EventPut,
				Rev:      etcdKv.ModRevision,
				Ver:      etcdKv.Version,
				Template: template,
			}
			select {
			case eventCh <- e:
			case <-done:
				return nil
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
		}
		for {
			var resp etcd.WatchResponse
			var ok bool
			select {
			case resp, ok = <-rch:
			case <-done:
				return nil
			}
			if !ok {
				if err := internalWatcher.Close(); err != nil {
					return errors.EnsureStack(err)
				}
				// We interpret a response with nil Error and 0 events as a server side cancellation
				// as described in https://github.com/etcd-io/etcd/blob/client/v3.5.1/client/v3/watch.go#L53
				//
				// "If the context "ctx" is canceled or timed out, returned "WatchChan" is closed,
				// and "WatchResponse" from this closed channel has zero events and nil "Err()"."
				if resp.Err() == nil && len(resp.Events) == 0 {
					return context.Canceled
				}
				// use new "nextRevision"
				internalWatcher = etcd.NewWatcher(client)
				rch = internalWatcher.Watch(ctx, prefix, watchOptions(nextRevision)...)
				continue
			}
			if err := resp.Err(); err != nil {
				return errors.EnsureStack(err)
			}
			for _, etcdEv := range resp.Events {
				ev := &Event{
					Key:      bytes.TrimPrefix(etcdEv.Kv.Key, []byte(trimPrefix)),
					Value:    etcdEv.Kv.Value,
					Rev:      etcdEv.Kv.ModRevision,
					Ver:      etcdEv.Kv.Version,
					Template: template,
				}
				if etcdEv.Type == etcd.EventTypePut {
					ev.Type = EventPut
				} else {
					ev.Type = EventDelete
				}
				select {
				case eventCh <- ev:
				case <-done:
					return nil
				}
			}
			nextRevision = resp.Header.Revision + 1
		}
	}() //nolint:errcheck

	return &etcdWatcher{
		eventCh: eventCh,
		done:    done,
	}, nil
}

// MakeEtcdWatcher returns a Watcher that uses the given event channel and done
// channel internally to deliver events and signal closure, respectively.
func MakeEtcdWatcher(eventCh chan *Event, done chan struct{}) Watcher {
	return &etcdWatcher{
		eventCh: eventCh,
		done:    done,
	}
}

// CheckType checks to make sure val has the same type as template, unless
// template is nil in which case it always returns nil.
func CheckType(template proto.Message, val interface{}) error {
	if template != nil {
		valType, templateType := reflect.TypeOf(val), reflect.TypeOf(template)
		if valType != templateType {
			return errors.Errorf("invalid type, got: %s, expected: %s", valType, templateType)
		}
	}
	return nil
}
