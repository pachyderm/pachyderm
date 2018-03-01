# Getting Your Data into Pachyderm

Data that you put (or "commit") into Pachyderm ultimately lives in an object
store of your choice (S3, Minio, GCS, etc.).  This data is content-addressed by
Pachyderm to build our version control semantics and is therefore not
"human-readable" directly in the object store.  That being said, Pachyderm
allows you and your pipeline stages to interact with versioned files like you
would in a normal file system.

## Jargon associated with putting data in Pachyderm

### "Data Repositories"

Versioned data in Pachyderm lives in repositories (again think about something
similar to "git for data").  Each data "repository" can contain one file,
multiple files, multiple files arranged in directories, etc.  Regardless of the
structure, Pachyderm will version the state of each data repository as it
changes over time.

### "Commits"

Regardless of the method you use to get data into Pachyderm, the mechanism that
is used to get data into Pachyderm is a "commit" of data into a data
repository. In order to put data into Pachyderm a commit must be "started" (aka
an "open commit").  Then the data put into Pachyderm in that open commit will
only be available once the commit is "finished" (aka a "closed commit").
Although you have to do this opening, putting, and closing for all data that is
committed into Pachyderm, we provide some convenient ways to do that with our
CLI tool and clients (see below). 

## How to get data into Pachyderm

In terms of actually getting data into Pachyderm via "commits," there are
a couple of options:

- The `pachctl` CLI tool: This is the great option for testing and for users
  who prefer to input data scripting.
- One of the Pachyderm language clients: This option is ideal for Go, Python,
  or Scala users who want to push data to Pachyderm from services or
  applications written in those languages. Actually, even if you don't use Go,
  Python, or Scala, Pachyderm uses a protobuf API which supports many other
  languages, we just haven’t built the full clients yet.

### `pachctl`

To get data into Pachyderm using `pachctl`, you first need to create one or
more data repositories to hold your data:

```sh
$ pachctl create-repo <repo name>
```

Then to put data into the created repo, you use the `put-file` command. Below
are a few example uses of `put-file`, but you can see the complete
documentation [here](../pachctl/pachctl_put-file.html). Note again, commits in
Pachyderm must be explicitly started and finished so `put-file` can only be
called on an open commit (started, but not finished). The `-c` option allows
you to start and finish a commit in addition to putting data as a one-line
command. 

Add a single file to a new branch:

```sh
# first start a commit
$ pachctl start-commit <repo> <branch>

# then put <file> at <path> in the <repo> on <branch>
$ pachctl put-file <repo> <branch> </path/to/file> -f <file>

# then finish the commit
$ pachctl finish-commit <repo> <branch>
```

Start and finish a commit while adding a file using `-c`:

```sh
$ pachctl put-file <repo> <branch> </path/to/file> -c -f <file> 
```

Put data from a URL:

```sh
$ pachctl put-file <repo> <branch> </path/to/file> -c -f http://url_path
```

Put data directly from an object store:

```sh
# here you can use s3://, gcs://, or as://
$ pachctl put-file <repo> <branch> </path/to/file> -c -f s3://object_store_url
```

Put data directly from another location within Pachyderm:

```sh
$ pachctl put-file <repo> <branch> </path/to/file> -c -f pfs://pachyderm_location
```

Add multiple files at once by using the `-i` option or multiple `-f` flags. In
the case of `-i`, the target file should be a list of files, paths, or URLs
that you want to input all at once:

```sh
$ pachctl put-file <repo> <branch> -c -i <file containing list of files, paths, or URLs>
```

Pipe data from stdin into a data repository:

```sh
$ echo "data" | pachctl put-file <repo> <branch> </path/to/file> -c
```

Add an entire directory or all of the contents at a particular URL (either
HTTP(S) or object store URL, `s3://`, `gcs://`, and `as://`) by using the
recursive flag, `-r`:

```sh
$ pachctl put-file <repo> <branch> -c -r <dir>
```

### Pachyderm Language Clients

There are a number of Pachyderm language clients.  These can be used to
programmatically put data into Pachyderm, and much more.  You can find out more
about these clients [here](../reference/clients.html).

## Deleting Data

Sometimes "bad" data gets committed to Pachyderm and you need a way to delete
it. There are a couple of ways to address this,  which one you use depends on
what exactly was "bad" about the data you committed and what's happened in the
system since you committed the "bad" data.

### Deleting The Head of a Branch

The simplest case is when you've just made a commit to a branch with some
incorrect information. This is bad because now the head of your branch is
incorrect and users who read from it are likely to be misled. It's also bad if
there are pipelines which take this branch as an input, because there are now
jobs running which will process incorrect input, and either output incorrect
output or error. To fix this you should use `delete-commit` like so:

```sh
$ pachctl delete-commit <repo> <branch-or-commit-id>
```

When you delete a commit several things will happen (all atomically):

- The commit will be deleted.
- Any branch that the commit was the head of will have its head set to the
  commit's parent. If the commit's parent is `nil` the branch's head will be set
  to `nil`.
- If the commit has an children (commits which it is the parent of) those
  children's parent will be set to the deleted commits parent. Again, if the
  deleted commit's parent is `nil` then the children commit's parent will be
  set to `nil`.
- Any jobs which were created due to this commit will be deleted (running jobs
  get killed). This includes jobs which don't directly take the commit as
  input, but are farther downstream in your DAG.
- Output commits from deleted jobs will also be deleted and all of these
  effects will apply to those commits as well.

### Deleting Non-Head Commits

Deleting commits is more complicated if you've committed more data to the
branch before realizing there was bad data that needs to be deleted. You can
still delete the commit as in the previous section, however, unless the later
commits overwrote or deleted the bad data it will still be present in the
children commits. *Deleting a commit does not modify its children.* In git
terms, `delete-commit` is equivalent to squashing a commit out of existence,
it's not equivalent to reverting a commit. The reason for this is that the
semantics of revert can get ambiguous if files being reverted have been
otherwise modified. Git's revert can leave you with a merge conflict to solve,
and merge conflicts don't work well with Pachyderm due to the shared nature of
the system and the size of the data being stored.

You can also delete the children commits, however they may have good data that
you don't want to delete. If that's the case you should do the following:

1. Start a new commit on the branch with `pachctl start-commit`.
2. Delete all bad files from the newly opened commit with `pachctl delete-file`.
3. Finish the commit with `pachctl finish-commit`.
4. Delete the initial bad commits and all children up to the newly finished
   commit which contain the bad data.

Depending on how you're using Pachyderm the final step may be optional. After
you finish the "fixed" commit the heads of all your branches will converge to
correct results as downstream jobs finish. However, deleting those commits
allows you to clean up your commit history and makes sure that no one will ever
access errant data when reading non-head commits.

### Deleting Sensitive Data

If the data you committed is bad because it's sensitive and you want to make
sure that nobody ever accesses it there's an extra step in addition to those
above. Pachyderm stores data in a content addressed way and when you delete
a file, or a commit it only deletes references to the underlying data, it
doesn't delete the actual data until it does a garbage collection. To truly
purge the data you must delete all references to it using the methods described
above, and then run a garbage collect with `pachctl garbage-collect`.
