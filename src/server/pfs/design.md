# PFS v2 Design Doc

PFS v2 is a new implementation of the Pachyderm File System (PFS).  Architecturally, the most notable difference between v1 and v2 is that in v1, we store commit information in a custom distributed data store, whereas in v2 we store the information in RethinkDB.

As you shall see, there are two major attributes of RethinkDB that significantly influence the design of PFS v2:

* First-class support for secondary indexes
* Lack of multi-document transactions

## Motivation

See Github issue [#411](https://github.com/pachyderm/pachyderm/issues/411).

## Schema

See [persist.proto](db/persist/persist.proto).

## The Clock System

In the previous implementation of PFS, each commit stores a pointer to its parent.  Constructing a file, therefore, involves following a chain of parent pointers.  This is fine when we used to store all commits in memory.  However, in this new state of the world where commits are stored in a database, we cannot afford to use the old paradigm because each pointer traversal will be one round trip to the database.

Instead, we use a [logical clock](https://en.wikipedia.org/wiki/Logical_clock) system to track the causal/parental relationships among commits.  The system is mostly equivalent to [vector clocks](https://en.wikipedia.org/wiki/Vector_clock), where each branch is the equivalent of a node in a distributed system, and each commit is the equivalent of an event.

Each commit carries a vector clock.  The rule for setting the vector clock is simple:

1. If the commit is the first commit on a new branch, we append a clock `(branch_name, 0)` to the vector.
2. Otherwise, the commit inherits its parent's vector clock, with the last component incremented by one.

For instance, this is a valid commit graph with three branches:

```
[(foo, 0)] -> [(foo, 1)] -> [(foo, 2)] -> [(foo, 2)]
           |
           -> [(foo, 0), (bar, 0)] -> [(foo, 0), (bar, 1)] -> [(foo, 0), (bar, 2)]
                                                           |
                                                           -> [(foo, 0), (bar, 1), (buzz, 0)]
```

Intuitively, commit A is commit B's ancestor if and only if one of the following two conditions holds:

1. Commit A's clock is a prefix of commit B's clock.  For instance, `[(foo, 2), (bar, 3)]` is an ancestor of `[(foo, 2), (bar, 3), (buzz, 4)]`.
2. Commit A's clock is the same as commit B's clock except that the last component is smaller.  For instance, `[(foo, 2), (bar, 3)]` is an ancestor of `[(foo, 2), (bar, 5)]`.

Therefore, by creating a database index on the clocks of the commits, we can find all commits in a given commit range in one query.  For instance, to find all commits between [(foo, 2)] and [(foo, 4), (bar, 5), (buzz, 6)], the query looks like this:

```
query = between(left=[(foo, 2)], right=[(foo, 4)])
      + between(left=[(foo, 4), (bar, 0)], right=[(foo, 4), (bar, 5)]) 
      + between(left=[(foo, 4), (bar, 5), (buzz, 0)], right=[(foo, 4), (bar, 5), (buzz, 6)])
```

## Commits and Diffs

Commits and Diffs are the two main structures in our database schema.  A commit is logically associated with many diffs.  Each diff in a commit corresponds to a file or a directory that's modified in the commit.

One of our core design principals is to minimize the number of DB round trips for common operations, notably GetFile and PutFile.  To that end, we designed the schema such that PutFile only involves one round trip to the database.  Consider the following example:

```Go
PutFile(commitID, fileName, "a bunch of bytes")
PutFile(commitID, fileName, "another bunch of bytes")
```

Each of these `PutFile` calls results in an upsert to the same `Diff` document.  The two calls know to modify the same document because we construct the Diff document ID by concatenating `commitID` and `fileName`.  The first call is going to create the document, while the second call is going to update the document by appending references to the new blocks (i.e. the new content written).

Note that a `DeleteFile` call also creates a diff.  This diff contains no references to blocks and instead only has a `Delete` field set to `true`.

## Directories

The most obvious approach to represent file system hierarchy is to have "directory documents" that store the names of the files they contain.  Since RethinkDB does not support multi-document transactions, this approach has the consistency issue wherein the database can be stuck in a state where a file exists, but its name has not been added to the "directory document" that should contain it.

In PFS, the file system hierarchy is indirectly represented by indexes.  Consider the following example:

```Go
PutFile(commitID, "/foo/bar/buzz", "a bunch of bytes")
```

Here we are writing a file located at `/foo/bar/buzz`.  This `PutFile` call will actually upsert three `Diff` documents: `/foo`, `/foo/bar`, and `/foo/bar/buzz`.

Each `Diff` document is then indexed into a `DiffParentIndex`, where a document's index key is the path of its parent.  For instance, `/foo/bar/buzz` will have the key `/foo/bar` in the index, and `/foo`'s key will be `/`.

Therefore, to get all files under a path such as `/foo/bar`, we just query the index with the key `/foo/bar`.

## GetFile and Deletion

Recall that PFS is a version-controlled file system similar in spirit to Git.  In PFS, one of the most common operations is to read a file in a certain commit range.  Consider this example:

```Go
PutFile(commit1, fileName, "foo")
PutFile(commit2, fileName, "bar")
PutFile(commit3, fileName, "buzz")

GetFile(fromCommit=commit1, toCommit=commit3, fileName)
```

The `GetFile` call should return "barbuzz", because that's what's been written since `fromCommit` (i.e. commit1) to `toCommit` (i.e. commit3). 

Intuitively, we can start streaming bytes to the client as soon as we find the first diff in the commit range for the given path.  We do this through the `DiffClockIndex`.  This index maps a diff to a key that contains the path and the clock of the diff.  Consider this example:

```Go
PutFile([(master, 2)], "/foo/bar", "a bunch of bytes")
PutFile([(master, 4)], "/foo/bar", "a bunch of bytes")
```

These will create two `Diff` documents that have the keys `["/foo/bar", [(master, 2)]]` and `["/foo/bar", [(master, 4)]]` respectively.  Now imagine that we want to read `/foo/bar` between `(master, 1)` and `(master 5)`.  The database query we issue looks like this:

```Go
query = Between(index=DiffClockIndex, left=["/foo/bar", [(master, 1)]], right=["/foo/bar", [(master, 5)]])
```

It's easy to see how the two diff documents will be returned from this query, in the order we want.

However, the situation is complicated in the face of deletions.  Consider this example:

```Go
PutFile(commit1, fileName, "foo")
PutFile(commit2, fileName, "bar")
DeleteFile(commit3, fileName)
PutFile(commit4, fileName, "buzz")

GetFile(fromCommit=commit1, toCommit=commit4, fileName)
```

Here we need to return `buzz`, because the file has been deleted in commit 3.  Therefore, we can't just start streaming from commit 2.

What we need is a way to stream only the blocks that are added since the most recent deletion.  To that end, we use RethinkDB's [`fold`](https://rethinkdb.com/api/javascript/fold/) operation to transform a series of diffs into a single diff that contains only the blocks we want.  The details are not important; just know that the `fold` happens inside of the database, so in the end we still only need to send one query before we can start streaming data.
