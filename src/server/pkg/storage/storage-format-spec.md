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
