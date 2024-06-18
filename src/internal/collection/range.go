package collection

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Hoist these consts so users don't have to import etcd
var (
	SortByCreateRevision = etcd.SortByCreateRevision
	SortByModRevision    = etcd.SortByModRevision
	SortByKey            = etcd.SortByKey
	SortNone             = etcd.SortNone

	SortAscend  = etcd.SortAscend
	SortDescend = etcd.SortDescend
)

type SortTarget = etcd.SortTarget
type SortOrder = etcd.SortOrder

// Options are the sort options when iterating through etcd key/values.
// Currently implemented sort targets are CreateRevision and ModRevision.
type Options struct {
	Target SortTarget
	Order  SortOrder
	// Limit is only implemented for postgres collections
	Limit int
}

// DefaultOptions are the default sort options when iterating through etcd
// key/values.
func DefaultOptions() *Options {
	return &Options{SortByCreateRevision, SortDescend, 0}
}

func listFuncs(opts *Options) (func(*mvccpb.KeyValue) etcd.OpOption, func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int) {
	var from func(*mvccpb.KeyValue) etcd.OpOption
	var compare func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int
	switch opts.Target {
	case SortByCreateRevision:
		switch opts.Order {
		case SortAscend:
			from = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMinCreateRev(fromKey.CreateRevision) }
		case SortDescend:
			from = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMaxCreateRev(fromKey.CreateRevision) }
		}
		compare = func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int {
			return int(kv1.CreateRevision - kv2.CreateRevision)
		}
	case SortByModRevision:
		switch opts.Order {
		case SortAscend:
			from = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMinModRev(fromKey.ModRevision) }
		case SortDescend:
			from = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMaxModRev(fromKey.ModRevision) }
		}
		compare = func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int {
			return int(kv1.ModRevision - kv2.ModRevision)
		}
	}
	return from, compare
}

func listRevision(ctx context.Context, c *etcdReadOnlyCollection, prefix string, limitPtr *int64, opts *Options, f func(*mvccpb.KeyValue) error) error {
	etcdOpts := []etcd.OpOption{etcd.WithPrefix(), etcd.WithSort(opts.Target, opts.Order)}
	var fromKey *mvccpb.KeyValue
	from, compare := listFuncs(opts)
	for {
		if fromKey != nil {
			etcdOpts = append(etcdOpts, from(fromKey))
		}
		resp, done, err := getWithLimit(ctx, c, prefix, limitPtr, etcdOpts)
		if err != nil {
			return err
		}
		kvs := getNewKeys(resp.Kvs, fromKey)
		for _, kv := range kvs {
			if strings.Contains(strings.TrimPrefix(string(kv.Key), prefix), indexIdentifier) {
				continue
			}
			if err := f(kv); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}
		}
		if done {
			return nil
		}
		if compare(resp.Kvs[0], resp.Kvs[len(resp.Kvs)-1]) == 0 {
			return errors.Errorf("revision contains too many objects to fit in one batch (this is likely a bug)")
		}
		fromKey = kvs[len(kvs)-1]
	}
}

func getNewKeys(respKvs []*mvccpb.KeyValue, fromKey *mvccpb.KeyValue) []*mvccpb.KeyValue {
	if fromKey == nil {
		return respKvs
	}
	for i, kv := range respKvs {
		if string(kv.Key) == string(fromKey.Key) {
			return respKvs[i+1:]
		}
	}
	return nil
}

func getWithLimit(ctx context.Context, c *etcdReadOnlyCollection, key string, limitPtr *int64, opts []etcd.OpOption) (*etcd.GetResponse, bool, error) {
	for {
		limit := atomic.LoadInt64(limitPtr)
		resp, err := c.etcdClient.Get(ctx, key, append(opts, etcd.WithLimit(limit))...)
		if err != nil {
			if status.Convert(err).Code() == codes.ResourceExhausted && limit > 1 {
				atomic.CompareAndSwapInt64(limitPtr, limit, limit/2)
				continue
			}
			return nil, false, errors.EnsureStack(err)
		}
		if len(resp.Kvs) < int(limit) {
			return resp, true, nil
		}
		return resp, false, nil
	}
}
