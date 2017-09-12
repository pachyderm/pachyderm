# Getting Your Data into Pachyderm

Data that you put (or "commit") into Pachyderm ultimately lives in an object store of your choice (S3, Minio, GCS, etc.).  This data is content-addressed by Pachyderm to buid our version control semantics and are therefore is not "human-readable" directly in the object store.  That being said, Pachyderm allows you and your pipeline stages to interact with versioned files like you would in a normal file system.   

## Jargon associated with putting data in Pachyderm

### "Data Repositories"

Versioned data in Pachyderm lives in repositories (again think about something similar to "git for data").  Each data "repository" can contain one file, multiple files, multiple files arranged in directories, etc.  Regardless of the structure, Pachyderm will version the state of each data repository as it changes over time.

### "Commits"

Regardless of the method you use to get data into Pachyderm, the mechanism that is used to get data into Pachyderm is a "commit" of data into a data repository. In order to put data into Pachyderm a commit must be "started" (aka an "open commit").  Then the data put into Pachyderm in that open commit will only be available once the commit is "finished" (aka a "closed commit"). Although you have to do this opening, putting, and closing for all data that is committed into Pachyderm, we provide some convient ways to do that with our CLI tool and clients (see below). 

## How to get data into Pachyderm

In terms of actually getting data into Pachyderm via "commits," there are a couple of options:

- The `pachctl` CLI tool: This is the great option for testing and for users who prefer to input data scripting.
- One of the Pachyderm language clients: This option is ideal for Go, Python, or Scala users who want to push data to Pachyderm from services or applications writtern in those languages. Actually, even if you don't use Go, Python, or Scala, Pachyderm uses a protobuf API which supports many other languages, we just haven’t built the full clients yet.

### `pachctl`

To get data into Pachyderm using `pachctl`, you first need to create one or more data repositories to hold your data:

```sh
$ pachctl create-repo <repo name>
```

Then to put data into the created repo, you use the `put-file` command. Below are a few example uses of `put-file`, but you can see the complete documentation [here](../pachctl/pachctl_put-file.html). Note again, commits in Pachyderm must be explicitly started and finished so `put-file` can only be called on an open commit (started, but not finished). The `-c` option allows you to start and finish a commit in addition to putting data as a one-line command. 

Add a single file to a new branch:

```sh
# first start a commit
$ pachctl start-commit <repo> <branch>

# then utilize the returned <commit-id> in the put-file request
# to put <file> at <path> in the <repo>
$ pachctl put-file <repo> <commit-id> </path/to/file> -f <file>

# then finish the commit
$ pachctl finish-commit <repo> <commit-id>
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

Add multiple files at once by using the `-i` option or multiple `-f` flags. In the case of `-i`, the target file should be a list of files, paths, or URLs that you want to input all at once:

```sh
$ pachctl put-file <repo> <branch> -c -i <file containing list of files, paths, or URLs>
```

Pipe data from stdin into a data repository:

```sh
$ echo "data" | pachctl put-file <repo> <branch> </path/to/file> -c
```

Add an entire directory by using the recursive flag, `-r`:

```sh
$ pachctl put-file <repo> <branch> -c -r <dir>
```

### Pachyderm language clients

There are a number of Pachyderm language clients.  These can be used to programmatically put data into Pachyderm, and much more.  You can find out more about these clients [here](../reference/clients.html).

