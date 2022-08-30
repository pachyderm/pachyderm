/*
Package index implements a hierarchical file system index.

Each element in the hierarchy is an [index.Index], which refers to data in object storage via a [chunk.DataRef].
The referenced data is either the contents of a single file or a list of other indices spanning a given path range, stored as a marshalled protos.
A hierarchy always has a single top-level index, which stored in e.g. a file set's additive field to indicate its contents.

The [index.Reader] and [index.Writer] automatically traverse intermediate index layers, allowing consumers to only interact with file indices.
The reader makes use of the hierarchical range representation to allow quickly skipping to sub-ranges.
The writer creates intermediate index levels automatically when accumulated index data (metadata, not underlying file data) reaches [index.DefaultBatchThreshold].

See the backwards compatibility section in the [chunk] documentation for some caveats that also apply to indices.

Note that file indices are inherently tied to a datum, with a default value for user-supplied files.
Low-level consumers should be aware of this and not assume that a given path will only appear once.
In general, due to our deletion mechanism, even files that are ultimately deleted by a composite file set
may still be present in earlier individual primitive file sets.
*/
package index
