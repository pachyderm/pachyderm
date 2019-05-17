This is an informal description of the storage format spec we will be defining
to help guide the development of Pachyderm's new storage layer. This is a work
in progress so the details are currently subject to change.

**Storage Format Basics:**

- The primitive for this storage format is the tar stream with UStar format headers.
- There are two basic types of tar entries: index and content.
  - Index entries index into other tar streams.
  - Content entries are standard tar entries of type file or directory.
- All tar streams are sorted with respect to the lexicographical order of the
  header names. Duplicate header names are allowed and are interpreted as
  a merge (described in the merge section).
- This storage format:
  - requires an implementation of a content-addressed chunk storage layer.
  - uses protocol buffers as the serialization format for indexes.

**Chunk Storage Layer:**

- The chunk storage layer should:
  - be able to convert arbitrary byte streams into content-addressed chunks and
    store them.
  - provide an interface for creating and accessing ranges within/across these
    byte streams.
- A reference to data within a chunk is stored as a DataRef protocol buffer
  message which includes the chunk's content-address(hash), offset, and size.
- Referencing a range across multiple chunks requires stringing together
  multiple of these single chunk references.
- The hash of the data may also be included if it does not match the chunk hash
  (the data is only part of the chunk).

```
message DataRef {
	Chunk chunk
	string hash
	int64 offset_bytes
	int64 size_bytes
}

message Chunk {
	string hash
}
```

**Index Entry:**

- An index entry contains an index into other tar streams.
- The header name of an index entry is the same as the first header in the
  indexed tar stream.
- Index entries may index a tar stream with only one tar entry or a range of
  tar entries.
    - Index entries that index one tar entry have headers with type flag set to
      'i' and have additional metadata about the single tar entry.
    - Index entries that index a range of tar entries have headers with type
      flag set to 'r' and have additional metadata about the range of tar
      entries.
- The index is stored within an Index protocol buffer message that is
  serialized into the content section of the index entry.

```
message Index {
	Range Range
	DataOp dataOp
}

message Range {
	string lastPath
	// other stuff (bloom filter, stats, etc.)
}

message DataOp {
	repeated DataRef dataRef
	Op op
	repeated Tag tags
}

enum Op {
	APPEND
	OVERWRITE
	DELETE
}

message Tag {
	int64 id
	int64 size_bytes
}
```

- The indexed tar stream is represented by a data operation which may contain
  a sequence of data references that are associated with the operation (there
  should be no data references for a delete). The sequence of data references
  should be stored in the order they should be read to get the final state of
  the underlying tar stream.
- A sequence of tags can be associated with a data operation to communicate the
  relative ordering of data between data operations during merge. The tags are
  ordered in correspondence with the data, such that reading the data
  references and tags together provides enough information to determine the
  mapping from tag to data.

**Merging:**

- A merge is needed when content for a path(s) is spread across multiple tar
  streams.
- The merge operation in general should apply the data operations of the
  path(s) that show up in multiple tar streams, and create a new data operation
  that represents the merged state of the path.
- A merge is represented in the storage format by a sequence of tar entries
  that have overlapping ranges.
  - The range for an index entry that indexes one tar entry and a content entry
    is simply the tar header name.
  - The range for an index entry of that indexes a range of tar entries is the
    range formed by the tar header name and the lastPath field (inclusive).
- Based on the additional metadata associated with certain index entries, it
  may make sense to merge certain files at read time without re-writing storage
  or to merge the data eagerly and re-write storage.

**Garbage Collection:**

We store two types of objects in object storage, content addressed objects and
semantically addressed objects. Content address objects are named based on
their content, semantically named objects have their names passed in and are
identifiers that users of the storage layer defines. In practice these will be
commit names although we may start using it for other identifiers as well. The
important differences between these two types of objects as it pertains to GC is:

- CA objects are referenced by SA objects, and by each other. They cannot be
  deleted by the user, they must be deleted by GC.
- SA objects aren't referenced by other objects, they are deleted by the user
  and this deletion orphans CA objects and allows them to be GCed.

For each object we store we will also store a counting bloom filter which
describes the transitive closure of the objects it references. Counting
bloom filters have the nice property that they can be added to, and
subtract from each other. Bloom filters are appealing for GC because they
are space efficient structures that sometimes return false positives, but
never return false negatives. That means that if the bloom filter says an
object is unreferenced we know it's safe to delete. By summing the bloom
filters of all the SA objects we can get a global bloom filter that
describes all of the referenced blocks in our cluster, we can keep this up
to date by adding the bloom filters of newly added blocks, and subtracting
the bloom filters of deleted blocks. One of the difficulties of with bloom
filters is that they can't be used to enumerate paths, they can only be
used to check a path that you already know about. That means that to
perform garbage collection we'll need to scan objects and check them
against our bloom filter to see if they're still referenced. I think this
can work fine as a process that is always passively running.

The other big difficulty with Pachyderm GC is race conditions. Because of
our de-dupe system, an object that is unreferenced and thus ready to be
GCed may at any point get re-referenced and thus cease to be ready for GC.
I think we can solve this with a generational system wherein each object
belongs to a generation, and we record a bloom filter that describes the
reference changes in a generation. By summing all the generations together
we can get a global bloom filter for a generation and figure out what's
safe to delete. This is best demonstrated with an example:

| gen | ops          | bloom     | bloom sum |
| --- | ------------ | ------    | --------- |
|   0 | +foo         | 100101    | 100101    |
|   1 | +bar         | 001011    | 101112    |
|   2 | -foo         | -100-10-1 | 001011    |

After generation 2, foo is ready to be deleted, which is reflected in the
fact that the bloom sum returns negative for foo's bloom filter (100101).
The tricky part is that if we now want to re-add foo in gen 3 we'll need
to make sure that the GC process for gen 2 has either already finished,
because if it hasn't, it might GC foo after we write it for gen 3, or
knows not to GC this object, because it's about to be re-referenced.
I think all of this is manageable, but it's definitely tricky.
