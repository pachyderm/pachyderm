package collection

import (
	"fmt"
	"sort"
	"sync/atomic"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Options struct {
	Target   etcd.SortTarget
	Order    etcd.SortOrder
	SelfSort bool
}

var DefaultOptions *Options = &Options{etcd.SortByCreateRevision, etcd.SortDescend, false}

type etcdIterator interface {
	next() ([]*mvccpb.KeyValue, error)
	done() bool
}

type iterator struct {
	c           *readonlyCollection
	prefix      string
	limitPtr    *int64
	opts        *Options
	compareFunc func(*mvccpb.KeyValue, *mvccpb.KeyValue) int
	last        bool
}

func newEtcdIterator(c *readonlyCollection, prefix string, limitPtr *int64, opts *Options) etcdIterator {
	if opts.SelfSort {
		iter := &selfSortRevisionIterator{iterator: iterator{c: c, prefix: prefix, limitPtr: limitPtr, opts: opts}}
		_, iter.compareFunc = iteratorFuncs(opts)
		return iter
	}
	iter := &revisionIterator{iterator: iterator{c: c, prefix: prefix, limitPtr: limitPtr, opts: opts}}
	iter.fromFunc, iter.compareFunc = iteratorFuncs(opts)
	return iter
}

func iteratorFuncs(opts *Options) (func(*mvccpb.KeyValue) etcd.OpOption, func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int) {
	var fromFunc func(*mvccpb.KeyValue) etcd.OpOption
	var compareFunc func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int
	switch opts.Target {
	case etcd.SortByCreateRevision:
		switch opts.Order {
		case etcd.SortAscend:
			fromFunc = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMinCreateRev(fromKey.CreateRevision) }
		case etcd.SortDescend:
			fromFunc = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMaxCreateRev(fromKey.CreateRevision) }
		}
		compareFunc = func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int {
			return int(kv1.CreateRevision - kv2.CreateRevision)
		}
	case etcd.SortByModRevision:
		switch opts.Order {
		case etcd.SortAscend:
			fromFunc = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMinModRev(fromKey.ModRevision) }
		case etcd.SortDescend:
			fromFunc = func(fromKey *mvccpb.KeyValue) etcd.OpOption { return etcd.WithMaxModRev(fromKey.ModRevision) }
		}
		compareFunc = func(kv1 *mvccpb.KeyValue, kv2 *mvccpb.KeyValue) int {
			return int(kv1.ModRevision - kv2.ModRevision)
		}
	}
	return fromFunc, compareFunc
}

type revisionIterator struct {
	iterator
	fromKey  *mvccpb.KeyValue
	fromFunc func(*mvccpb.KeyValue) etcd.OpOption
}

func (iter *revisionIterator) next() ([]*mvccpb.KeyValue, error) {
	opts := []etcd.OpOption{etcd.WithPrefix(), etcd.WithSort(iter.opts.Target, iter.opts.Order)}
	if iter.fromKey != nil {
		opts = append(opts, iter.fromFunc(iter.fromKey))
	}
	resp, done, err := getWithLimit(iter.c, iter.prefix, iter.limitPtr, opts)
	if err != nil {
		return nil, err
	}
	kvs := getNewKeys(resp.Kvs, iter.fromKey)
	if done {
		iter.last = true
		return kvs, nil
	}
	if iter.compareFunc(resp.Kvs[0], resp.Kvs[len(resp.Kvs)-1]) == 0 {
		return nil, fmt.Errorf("revision contains too many objects to fit in one batch (this is likely a bug)")
	}
	iter.fromKey = kvs[len(kvs)-1]
	return kvs, nil
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

func (iter *revisionIterator) done() bool {
	return iter.last
}

type selfSortRevisionIterator struct {
	iterator
	kvs []*mvccpb.KeyValue
}

func (iter *selfSortRevisionIterator) next() ([]*mvccpb.KeyValue, error) {
	opts := []etcd.OpOption{etcd.WithFromKey(), etcd.WithRange(endKeyFromPrefix(iter.prefix))}
	fromKey := iter.prefix
	iter.kvs = []*mvccpb.KeyValue{}
	for {
		resp, done, err := getWithLimit(iter.c, fromKey, iter.limitPtr, opts)
		if err != nil {
			return nil, err
		}
		if fromKey == iter.prefix {
			iter.kvs = append(iter.kvs, resp.Kvs...)
		} else {
			iter.kvs = append(iter.kvs, resp.Kvs[1:]...)
		}
		if done {
			iter.last = true
			break
		}
		fromKey = string(iter.kvs[len(iter.kvs)-1].Key)
	}
	if iter.opts.Order == etcd.SortAscend {
		sort.Sort(iter)
	} else {
		sort.Sort(sort.Reverse(iter))
	}
	return iter.kvs, nil
}

func endKeyFromPrefix(prefix string) string {
	// Lexicographically increment the last character
	return prefix[0:len(prefix)-1] + string(byte(prefix[len(prefix)-1])+1)
}

func (iter *selfSortRevisionIterator) done() bool {
	return iter.last
}

func (iter *selfSortRevisionIterator) Len() int {
	return len(iter.kvs)
}

func (iter *selfSortRevisionIterator) Less(i, j int) bool {
	return iter.compareFunc(iter.kvs[i], iter.kvs[j]) < 0
}

func (iter *selfSortRevisionIterator) Swap(i, j int) {
	t := iter.kvs[i]
	iter.kvs[i] = iter.kvs[j]
	iter.kvs[j] = t
}

func getWithLimit(c *readonlyCollection, key string, limitPtr *int64, opts []etcd.OpOption) (*etcd.GetResponse, bool, error) {
	for {
		limit := atomic.LoadInt64(limitPtr)
		resp, err := c.etcdClient.Get(c.ctx, key, append(opts, etcd.WithLimit(limit))...)
		if err != nil {
			if status.Convert(err).Code() == codes.ResourceExhausted && limit > 1 {
				atomic.CompareAndSwapInt64(limitPtr, limit, limit/2)
				continue
			}
			return nil, false, err
		}
		if len(resp.Kvs) < int(limit) {
			return resp, true, nil
		}
		return resp, false, nil
	}
}
