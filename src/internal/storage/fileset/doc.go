/*
Package fileset implements Pachyderm's file system.
The contents of a commit are made up of one or more file sets.
The underlying file set objects are ephemeral by default, and unreferenced objects (see [track.Tracker]) will be garbage collected.
Operations that involve creating and interacting with a file set should be done inside a [renew.StringSet] to handle keep alive behavior automatically.

Other packages will use the [fileset.FileSet] interface, which allows iteration over a lexicographically sorted collection of files.

The underlying file set objects themselves are either [fileset.Primitive] or [fileset.Composite], an ordered list of primitive file sets.
File sets are represented by random UUIDs
A Primitive file set represents a coherent set of changes to a file system, consisting of deleting all files in [Primitive.Deletive] followed by appending all files in [Primitive.Additive].
Overwriting a file is represented by deleting and appending in the same file set.
A composite file set represents the result of applying all of its underlying primitive file sets in order.

See the [index] package for details about the format.
*/
package fileset
