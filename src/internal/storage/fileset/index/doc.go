/*
Package index implements a hierarchical file system index.

Each element in the hierarchy is an [index.Index], which refers to data in object storage via a [chunk.DataRef].
The referenced data is either the contents of a single file or a list of other indices spanning a given path range, stored as a marshalled protos.
A hierarchy always has a single top-level index, which stored in e.g. a file set's additive field to indicate its contents.

The [index.Reader] and [index.Writer] automatically traverse intermediate index layers, allowing consumers to only interact with file indices.
The reader makes use of the hierarchical range representation to allow quickly skipping to sub-ranges, directories, or individual files.
The writer creates intermediate index levels automatically when accumulated index data (metadata, not underlying file data) reaches [index.DefaultBatchThreshold].

Note that file indices are inherently tied to a datum, with a default value for user-supplied files.
While it's an error for a commit's file system to contain the same file path under multiple datums,
it's valid for one of its component file sets to exhibit this duplication.
Low-level consumers should be aware of this and not assume that a given path will only appear once.

See the backwards compatibility section in the [chunk] documentation for some caveats that also apply to indices.
*/
package index
