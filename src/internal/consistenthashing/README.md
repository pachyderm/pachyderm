# Consistent Hashing
The consistent hashing library implements a basic 
[consistent hashing](https://ably.com/blog/implementing-efficient-consistent-hashing) `Ring` that allows users to manage 
locks in ETCD across multiple processes modelled as `Nodes` while handling a dynamic number of Nodes and lock rebalancing.

A ring should be constructed by calling `New()`. Once instantiated, nodes can be added to ring with `AddNode()`.
Once one or more node has been added, a user can `lock` keys by calling `Lock()` and release them with `Unlock()`.

If a node is added or removed from a ring, each ring instance determines whether it needs to rebalance the existing 
locks in the ring across all nodes using a hash function `HashFn()`.

## Uses
Consistent hashing rings are used 

## Rings
In a normal use case, a ring instance is created by each process. In our case that is a container running on Kubernetes.
Ring instances that share a `Prefix` watch keys in ETCD that share that prefix for changes by launching an ETCD watcher
in a separate goroutine. Users must also pass in a hash function to a ring. A sane default is `crc32.ChecksumIEEE()`.

## Nodes
A Ring may have 0 or more nodes. Nodes added to a ring share a prefix and are tracked by each ring instance watching
that prefix. Once a new node is detected, rings that share its prefix add its metadata to their `nodeInfos`. This 
includes the node's `Id` and its `Hash`, which is used by the consistent hashing algorithm to determine how to distribute
keys and to rebalance them if a node is added or removed. The ring that creates a node tracks the `locks` that will 
be associated to that node via the ring's hash function.

Adding a node starts a goroutine that refreshes the node's lease in ETCD which acts as a liveness check. If the lease expires,
the node is removed from the ring. Deleting a node shuts that goroutine down and also deletes the node's Id in ETCD. 
While adding is idempotent, delete is not. Delete will not delete a node that was added by a remote ring instance. For 
instance if two processes, A and B, each create a ring instance over prefix "p", then add a node, if A tries to delete 
the node added by B, the call to DeleteNode() will return an error.

## Locks
Attempting to lock a key with a ring instance is a blocking operation. When attempting to lock a key, the ring hashes 
the key to determine which node associates to the key. If that node is local to the ring instance, the ring calls lock
on a mutex for that key. Otherwise, it re-polls the output of the hash function until the call to Lock() is cancelled or
until the node that the key associates to is local to the ring instance. This can happen when nodes are removed from a
ring. When new nodes are added, each ring instance rebalances the locks across its local nodes by iterating through all locks
and checking whether those locks still associate to the same nodes.

Unlock also determines whether a lock associates to a node that is local to the ring instance. If so, the mutex is 
unlocked and removed from the set of locks.
