package server

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

func kubeEventTail(ctx context.Context, coreV1 corev1.CoreV1Interface, namespace string, etcdPrefix string, etcdClient *etcd.Client) {
	// kubernetes's fake ClientSet used in our unit tests doesn't support the RESTClient.
	// in unit tests we don't run kube event tail.
	if coreV1.RESTClient() == nil || reflect.ValueOf(coreV1.RESTClient()).IsNil() {
		log.Info(ctx, "skip kube event tail. kubernetes client not fully configured.")
		return
	}
	backoff.RetryUntilCancel(ctx, func() (retErr error) { //nolint:errcheck
		lock := dlock.NewDLock(etcdClient, path.Join(etcdPrefix, "pachd-kube-events"))
		ctx, err := lock.Lock(ctx)
		if err != nil {
			return errors.Wrap(err, "locking pachd-kube-events lock")
		}
		defer errors.Invoke1(&retErr, lock.Unlock, ctx, "error unlocking")
		lw := cache.NewListWatchFromClient(coreV1.RESTClient(), "events", namespace, fields.Everything())
		r := cache.NewReflector(lw, &v1.Event{}, &eventStore{ctx: ctx}, 0)
		r.Run(ctx.Done())
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error in kube event tail; will retry", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

func (s *eventStore) logEvent(e *v1.Event) {
	if e == nil {
		return
	}
	msg := "event without message"
	if e.Message != "" {
		msg = e.Message
	}

	objRef := func(o *v1.ObjectReference) zapcore.ObjectMarshaler {
		return zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			if o.Namespace != e.Namespace {
				enc.AddString("namespace", o.Namespace)
			}
			enc.AddString("name", o.Kind+"/"+o.Name)
			if p := o.FieldPath; p != "" {
				enc.AddString("fieldPath", p)
			}
			return nil
		})
	}
	log.Info(s.ctx, msg,
		zap.Object("kubeEvent", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			var errs error
			enc.AddString("namespace", e.ObjectMeta.Namespace)
			enc.AddString("name", e.ObjectMeta.Name)
			if err := enc.AddObject("involvedObject", objRef(&e.InvolvedObject)); err != nil {
				errors.JoinInto(&errs, fmt.Errorf("marshal involvedObject: %w", err))
			}
			if e.Reason != "" {
				enc.AddString("reason", e.Reason)
			}
			src := e.Source
			if src.Component != "" {
				enc.AddString("source.component", src.Component)
			}
			if src.Host != "" {
				enc.AddString("source.host", src.Host)
			}
			if t := e.EventTime.Time; !t.IsZero() {
				enc.AddTime("eventTime", t)
			}
			if t := e.FirstTimestamp.Time; !t.IsZero() {
				enc.AddTime("firstTimestamp", t)
			}
			if t := e.LastTimestamp.Time; !t.IsZero() {
				enc.AddTime("lastTimestamp", t)
			}
			if f, l := e.FirstTimestamp, e.LastTimestamp; !f.IsZero() && !l.IsZero() && l.Sub(f.Time) > time.Second {
				enc.AddDuration("ongoingDuration", l.Sub(f.Time))
			}
			if e.Count > 0 {
				enc.AddInt32("count", e.Count)
			}
			if s := e.Series; s != nil {
				enc.AddInt32("series.count", s.Count)
				enc.AddTime("series.lastObservedTime", s.LastObservedTime.Time)
			}
			if a := e.Action; a != "" {
				enc.AddString("action", a)
			}
			if e.Related != nil {
				if err := enc.AddObject("related", objRef(e.Related)); err != nil {
					errors.JoinInto(&errs, fmt.Errorf("marshal related object: %w", err))
				}
			}
			if rc := e.ReportingController; rc != "" {
				enc.AddString("reportingController", rc)
			}
			if ri := e.ReportingInstance; ri != "" {
				enc.AddString("reportingInstance", ri)
			}
			if t := e.Type; t != "" && t != "Normal" {
				enc.AddString("type", t)
			}
			if errs != nil {
				enc.AddString("marshalErrors", errs.Error())
			}
			return nil
		})),
	)
}

// eventStore is a cache.Store that logs events as they are added.
type eventStore struct {
	ctx context.Context
}

// Add implements cache.Store.
func (s *eventStore) Add(obj interface{}) error {
	e, ok := obj.(*v1.Event)
	if !ok {
		return errors.New("non-event object received")
	}
	s.logEvent(e)
	return nil
}

// Update implements cache.Store.
//
// NOTE(jrockway): Events are actually updated; this is relevant when looking at fields like
// LastTimestamp and Count.  We can't change logs we've already written, so we treat them as brand
// new events.  What this means is that if you ever wanted to implement a feature like "don't show
// cached events on startup" and suppress events with ObjectMeta.CreationTime before the program
// started up, you'll also suppress events that are happening now but are updates to an event
// resource that was created before the program started.  So some care is necessary, and is why the
// feature currently doesn't exist.
func (s *eventStore) Update(obj interface{}) error {
	return s.Add(obj)
}

// Delete implements cache.Store.
func (*eventStore) Delete(obj interface{}) error { return nil }

// Replace implements cache.Store.
func (s *eventStore) Replace(objs []interface{}, unusedResourceVersion string) error {
	var result error
	for _, obj := range objs {
		if err := s.Add(obj); err != nil {
			result = err
		}
	}
	return result
}

// We only implement cache.Store for cache.Reflector, and cache.Reflector does not call List/Get methods.
func (*eventStore) Resync() error       { return nil }
func (*eventStore) List() []interface{} { return nil }
func (*eventStore) ListKeys() []string  { return nil }
func (*eventStore) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return nil, false, errors.New("unimplemented")
}
func (*eventStore) GetByKey(key string) (item interface{}, exists bool, err error) {
	return nil, false, errors.New("unimplemented")
}
