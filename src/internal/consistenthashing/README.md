# Consistent Hashing
The consistent hashing library implements a basic 
[consistent hashing](https://ably.com/blog/implementing-efficient-consistent-hashing) `Ring` that allows users to manage 
locks in ETCD across multiple processes modelled while handling a dynamic number of Nodes and lock rebalancing.

Users should call `WithRing()` to create a scope where a ring can be used within a provided callback function. 
Once instantiated, a user can `lock` keys in the callback function by calling `Lock()` and release them with `Unlock()`.

If a node is added or removed from a ring, each ring instance determines whether it needs to rebalance its own locks 
using a hash function `HashFn()`.

## Uses
Consistent hashing rings are used in cases where multiple nodes should not attempt to use a given resource, but the
number of resources and nodes is dynamic. Because nodes can be removed or added at any time, using consistent hashing
enables a user to minimize the number of resources that need to be reassigned when the number of nodes changes. It also
provides an elegant way to model resources without having to track all resources across all nodes.

## Rings
In a normal use case, a ring instance is created by each process. In our case that is a container running on Kubernetes.
Ring instances that share a `Prefix` watch keys in ETCD that share that prefix for changes by launching an ETCD watcher
in a separate goroutine.

A Ring has 1 or more `members`. Each ring instance creates a local member called a `node`. Once a new member is detected, 
rings that share its prefix add its metadata to their `members`. This includes the member's `Id` and its `Hash`, which 
is used by the consistent hashing algorithm to determine how to distribute keys and to rebalance them if a member is 
added or removed. Each ring instance also tracks the `locks` that are associated to its node via the ring's hash function.
Nodes run a goroutine that refreshes their lease in ETCD which acts as a liveness check. If a lease expires,
a delete event is generated, triggering each ring instance to remove that member from their list of members. Deleting a 
ring instance shuts down its watch and refresh goroutines.

## Locks
Attempting to lock a key is a blocking operation. When attempting to lock a key, the ring hashes the key to determine 
whether its node associates to the key. If so, the ring calls lock on a mutex for that key. Otherwise, it re-polls the 
output of the hash function until the call to Lock() is cancelled or until its node associates to key. This can happen 
when nodes are removed from a ring. When new nodes are added, each ring instance rebalances its node's locks by 
rehashing the lock keys and checking whether those locks still associate to the same nodes.

Unlock also determines whether a lock associates to the ring instance's node. If so, the mutex is 
unlocked and removed from the set of locks.
