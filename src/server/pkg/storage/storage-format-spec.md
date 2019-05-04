This is an informal description of the storage spec we will be defining to help guide the development of Pachyderm's new storage layer. This is a work in progress so the details are currently subject to change.

**Storage Format Basics:**

- The primitive for this storage format is the tar stream with UStar format headers.
- There are two basic types of tar entries: index and content.
  - Index tar entries have headers with type flag set to 'i'.
  - Content tar entries have headers with the standard type flags corresponding to whether they are files or directories.
- All tar streams are sorted with respect to the lexicographical order of the header names. Duplicate header names are allowed and are interpreted as a merge (described in the merge section).
- This storage format is designed with an immutable object storage backend in mind and a chunk storage layer that is able to convert arbitrary byte streams into content addressed chunks and provides an interface for accessing/modifying ranges within/across these chunks.
- This storage format uses protocol buffers as the serialization format for additional metadata that is not stored in the tar headers.

**Chunk Storage:**

- A reference to data within a chunk is stored as a DataRef protocol buffer message which includes the chunk hash, offset, and size.
- Referencing a range across multiple chunks requires stringing together multiple of these single chunk references. 
- The hash of the data may also be included if it does not match the chunk hash (the data is only part of the chunk).

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

**Index:**

- The content of an index tar entry contains metadata about other tar streams. 
- The header name in an index tar entry is the same as the first header in the referenced tar stream. 
- There are two types of index entries: File and FileRange.
  - Index entries of type File indicate a direct reference to a content entry for a file and store additional information about the referenced file. 
  - Index entries of type FileRange indicate a reference to a range of files (can be a range of index and/or content entries) and store additional information about the file range. 
  - The additional information for each of these types are stored within a Metadata protocol buffer message that is serialized into the content section of the tar entry. 

```
message Metadata {
	// A non-empty fileRange field indicates that this entry is of type FileRange (vice versa for type File).
	FileRange fileRange
	DataOp dataOp
}

message FileRange {
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

- The referenced tar stream is represented by a data operation which may contain a sequence of data references that are associated with the operation (there should be no data references for a delete). The sequence of data references should be stored in the order they should be read to get the final state of the underlying tar stream.
- A sequence of tags can be associated with a data operation to communicate the relative ordering of data between data operations during merge. The tags are ordered in correspondence with the data, such that reading the data references and tags together provides enough information to determine the mapping from tag to data.

**Content:**

- A content tar entry is a standard tar entry (header followed by file content).

**Merging:**

- A merge is needed when content for a path(s) is spread across multiple tar streams.
- The merge operation in general should apply the data operations of the path(s) that show up in multiple tar streams, and create a new data operation that represents the merged state of the path.
- A merge is represented in the storage format by a sequence of tar entries that have overlapping ranges.
  - The range for an index entry of type File and content entry is simply the tar header name.
  - The range for an index entry of type FileRange is the range formed by the tar header name and the lastPath field (inclusive).
- Based on the additional metadata associated with certain index entries, it may make sense to merge certain files at read time without re-writing storage or to merge the data eagerly and re-write storage.
