/*
Package chunk is Pachyderm's interface with object storage.

A chunk is the basic unit of storage that PFS uses for arbitrary data.
Chunks are identified by the hash of the underlying data, which, along with content-defined chunking, is the mechanism of PFS deduplication.
Content-defined chunking is implemented by [chunk.ComputeChunks].
This package also supports a second way of chunking data, [chunk.Batcher], used for small files where chunk overhead outweighs deduplication benefits.

# Backwards compatibility

Some implementation details might not make sense based on the current code (e.g. the [chunk.Ref.Edge] check in [chunk.StableDataRefs]).
The algorithms for file and index chunking have changed in 2.x, and we must support previously-written data.
Here are some examples of conditions in past data which current code will not generate:
  - a file that is split across multiple chunks may share some of them with other files
	(in current code, a file split across chunks will be the only file in those chunks)
  - related, even small files may be split across multiple chunks
  - an index range data reference may start part way through a chunk
*/
package chunk
