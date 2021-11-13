# Collections

Collections are a library for storing protobufs in a persistent database with a simple interface for reading, writing, and watching for changes.  There are two existing implementations, one for etcd and one for postgres.

A collection can be constructed via `NewEtcdCollection` or `NewPostgresCollection`, depending on the backend to be used.  This function must be given:
 * a template protobuf `Message`, which will be clones for use in serialization and deserialization
 * a set of fields to create secondary indexes for (used by `GetByIndex` and `WatchByIndex`)

## API

The collection API is split into two means of accessing a collection - `ReadOnly` and `ReadWrite`.  A `ReadOnly` collection does not make any consistency guarantees, while a `ReadWrite` collection is guaranteed to do all reads and writes within a single transaction, although this comes with trade-offs.

### ReadOnly

A `ReadOnlyCollection` is constructed with a context that can be used for cancellation.

Operations:
 * `Get`, `GetByIndex` - point reads
 * `List` - stream the entire collection with some sorting options
 * `Count` - get the number of items in the collection
 * `Watch`, `WatchF` - stream the entire collection, then get notified of changes until canceled
 * `WatchOne`, `WatchOneF` - stream changes to a single item in the collection

 `EtcdReadOnlyCollection` additionally supports:
 * `TTL` - get the time-to-live of a given item in the collection
 * `ListRev` - stream the entire collection alongside the etcd revision of each item
 * `WatchByIndex` - watch the entire collection based on an index (unused)

### ReadWrite

A `ReadWriteCollection` is constructed with a transaction (`STM` for etcd, `sqlx.Tx` for postgres) which is used to isolate any operations until the transaction is finished.  The transaction itself may reattempt its callback multiple times in case the transaction is invalidated by other clients to the database.  It is important to note that you cannot preserve transactionality between an `STM` and a `sqlx.Tx`, so any code that must change things in both etcd and postgres must be aware of this (and generally, such code should be avoided).

Operations:
 * `Get` - same as in `ReadOnlyCollection`, but guaranteed to be consistent with the state of the transaction
 * `Put` - create or overwrite a given item in the collection
 * `Update` - get a given item from the collection, modify it in a callback, then write it out afterwards
 * `Upsert` - same as `Update` except it will continue with the callback and insert if the item does not exist already
 * `Create` - create a given item in the collection, error if it already exists
 * `Delete` - remove a given item from the collection, error if it does not exist
 * `DeleteAll` - remove all items from the collection

`EtcdReadWriteCollection` additional supports:
 * `TTL` - get the time-to-live of a given item in the collection
 * `PutTTL` - write an item to the collection that will be removed after the given time-to-live

## Implementation

## Etcd

TODO

### STM

The `STM` can be constructed via `NewSTM` (or `NewDryrunSTM`, which will not commit changes).  This code is copied from the etcd library in order to extend it.  It is a relatively common problem that etcd transactions will get too large and be rejected by etcd.  While the etcd config can be adjusted to increase the limit on transaction size, this can only do so much.  For any arbitrarily-large transactions, you may need to find a way to break it up into pieces of a controlled size.

## Postgres

The postgres implementation is written to mirror the functionality of the etcd implementation rather than to provide an ORM-style table of the protobuf fields inside a database.

### sqlx.Tx

The `sqlx.Tx` can be constructed via `NewSQLTx`.  This will occupy a sql session for its lifetime.  Constructing tehe `sqlx.Tx` will result in a `BEGIN;` at the start, and a `COMMIT;` at the end of the callback, unless an error occurs, in which case a `ROLLBACK;` will be issued.  This will reattempt up to three times in the case of a commit error.

### Tables

A table will be created for each postgres collection.  Construction of the collection will check that the table exists, as well as all relevant triggers and indexes.  If any are missing, they will be created.

The `model` struct exists to mirror the columns in the database for reading results from `sqlx` into a struct:
 * `CreatedAt` (golang) -> `createdat` (postgres) - the timestamp at which the item was created 
 * `UpdatedAt` (golang) -> `updatedat` (postgres) - the timestamp at which the item was last modified
 * `Version` (golang) -> `version` (postgres) - the version of pachyderm which serialized the protobuf
 * `Key` (golang) -> `key` (postgres) - the primary key for the item
 * `Proto` (golang) -> `proto` (postgres) - the item's serialized protobuf

 Additionally, columns will be created for each indexed field, following the pattern `idx_<field>` - these are only used for indexing purposes and are never actually read back out by a client, the user can get that information from the protobuf.

### Triggers

Two triggers exist on each postgres collection table, `update_modified_trigger` and `notify_watch_trigger`.

`update_modified_trigger` will modify the row before any write to update the `updatedat` column to the current time.

`notify_watch_trigger` will issue `NOTIFY` commands to one or several channels in case any clients are running watches on the collection.  For table-wide watches, the channel is straightforwardly `pwc_<table>`.  For point-watches, the trigger will also notify one channel for each index on the collection.  The channel name must depend on the value of the field being indexed so that clients do not have to discard O(n) events per interesting event.  Unfortunately, channel names must be valid postgres identifiers which puts several limitations on the channel name, so the value must be hashed.  This still leaves the possibility open for a collision in the channel name, so the client must still check that the event belongs to the watch in question (see [Listeners](#listeners) for details).

The payload on the channel includes several fields:
 * the field name of the index being notified
 * the value of the index being notified
 * the primary key of the row
 * the postgres operation that caused the trigger (`'INSERT'`, `'UPDATE'`, `'DELETE'`)
 * the timestamp of the transaction that caused the trigger
 * a base64-encoded serialized protobuf of the item

### Listeners

In order to provide `watch` semantics, a postgres connection must be dedicated to listening for notifications.  This functionality is not available on `sqlx` connections, so the `pq` library is used instead.  This is the same library that provides the underlying connection for `sqlx`, although we don't get much benefit from this.  A `PostgresListener` must be passed to any collection at construction time so that watches can be performed.  The `PostgresListener` lazily connect for the first `LISTEN` operation, and all `LISTEN`s should be multiplexed through a single connection.

Each `Watch` operation will listen on a postgres channel according to the watch type:
 * `Watch`, `WatchF` - listens to the table-wide `pwc_<table>` channel
 * `WatchOne`, `WatchOneF`, `WatchByIndex` - listens to a hashed channel `pwc_<table>_<hash>` based on the index key

 Once the `LISTEN` has returned, events must be buffered until the user has read out the current state of objects in the collections.  The `Watch` operation will then issue a `Get` or `List` operation to load the current state and keep track of the last modified timestamp.  The existing items will be provided to the `Watch` as `EventPut` events, and once these are done, the buffered events will be forwarded to the `Watch`, filtering out any that arrived before loading the existing items.

 Events that arrive from the `LISTEN` may not belong to the `Watch` due to hash collisions or a race condition in how we set up the `Watch`.  Therefore, the payload must be parsed to determine if the `Watch` is interested in the event, by comparing the index field and value to the `Watch`'s parameters, and potentially filtering out 'Delete' or 'Put' events, or events that arrived before the `Watch` was ready.
