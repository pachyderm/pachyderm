package drive

import (
	"context"
	"fmt"
	"strings"

	v3 "github.com/coreos/etcd/clientv3"
)

// newSequentialKV allocates a new sequential key <prefix>/nnnnn with a given
// value.  Note: a bookkeeping node __<prefix> is also allocated.
// It returns the sequential key created.
func (d *driver) newSequentialKV(ctx context.Context, prefix, val string) (string, error) {
	resp, err := d.etcdClient.Get(ctx, prefix, v3.WithLastKey()...)
	if err != nil {
		return "", err
	}

	// add 1 to last key, if any
	newSeqNum := 0
	if len(resp.Kvs) != 0 {
		fields := strings.Split(string(resp.Kvs[0].Key), "/")
		_, serr := fmt.Sscanf(fields[len(fields)-1], "%d", &newSeqNum)
		if serr != nil {
			return "", serr
		}
		newSeqNum++
	}
	newKey := fmt.Sprintf("%s/%016d", prefix, newSeqNum)

	// base prefix key must be current (i.e., <=) with the server update;
	// the base key is important to avoid the following:
	// N1: LastKey() == 1, start txn.
	// N2: new Key 2, new Key 3, Delete Key 2
	// N1: txn succeeds allocating key 2 when it shouldn't
	baseKey := "__" + prefix

	// current revision might contain modification so +1
	cmp := v3.Compare(v3.ModRevision(baseKey), "<", resp.Header.Revision+1)
	reqPrefix := v3.OpPut(baseKey, "")
	reqnewKey := v3.OpPut(newKey, val)

	txn := d.etcdClient.Txn(context.TODO())
	txnresp, err := txn.If(cmp).Then(reqPrefix, reqnewKey).Commit()
	if err != nil {
		return "", err
	}
	if !txnresp.Succeeded {
		return d.newSequentialKV(ctx, prefix, val)
	}
	return newKey, nil
}
