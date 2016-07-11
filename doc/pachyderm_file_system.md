Pachyderm File System (PFS)
===========================

Pachyderm File System - a version controlled file system for big data.

## Components

PFS has 4 basic primitives:

- File
- Commit
- Repo
- Block

Each of these is simple, and understanding all of them provides a good tour of PFS.

## Data as Files

Pachyderm File System (PFS) allows you to store arbitrary data in files. These files can be as large as you'd like, and store any kind of information.

We wanted to use an interface to data that is familiar to everyone. Reading/writing data to a file is as familiar as you get.

Doing this on big data sets gets interesting, but having a simple underlying interface makes interacting with the data more intuitive, and more easily accessible to developers no matter what their language of choices.

## Versioning

PFS is very git like. A data set is compromised of many `File`s, which constitutes a `Repo`. 

In PFS you version your data w `Commit`s. By versioning your data, you can:

- reproduce any input or output for your processing, which in turn enables ...
- collaborating w your peers on a data set

[Reproducibility and Collaboration](https://pachyderm.io/dsbor.html) are things we care a lot about.

We store each commit only as the data that changed from the prior commit. This is a concept borrowed from git. Storing your data this way also allows us to enable [Incrementality](https://pachyderm.io/dsbor.html).

## Files vs Blocks

Under the hood, we store your files in sets of `Block`s. These are smaller (usually ~8MB) chunks of your file. By storing your data in smaller chunks, we can more efficiently read and write your data in parallel.

`Block`s also determine the smallest indivisible chunk of your data. When performing a `map` job, each `File` is seen by multiple containers. Each container sees one or more `Block`s of a file.

## Block Delimiters

For certain data types (binary blobs or JSON objects), making sure that your data is divided _correctly_ into indivisible chunks is important. Doing this with PFS is straightforward. 

### Default

By default, data is line delimited and a single Block consists of 8MBs worth of lines.

### JSON

To make sure your JSON data is delimited correctly, just make sure the file in question has a `.json` suffix. This tells PFS that the data being stored is JSON, and each `Block` consists of whole JSON objects and can be up to 8MB

### Binary Data

Since binary data doesn't always have a static size, and can be quite large, delimiting binary data works a bit different.

This time, each write you do to a file (either via the CLI or literally a write syscall to the PFS mount) will determine a whole `Block` no matter what the size.

To require PFS to delimit blocks in this fashion, make sure your file as the `.bin` suffix.

## PFS I/O

What happens when you read or write to PFS?

### Storage

PFS is backed by an object store of your choosing (usually S3 or GCS). This allows for highly redundant consistent storage.

Each block of each of your files is content addressed and uploaded to your object store. This gains us de-duplication of the data.

Additionally, because each commit only contains deltas of blocks that were written, all data stored by PFS is immutable.

### Writing

You can never write to a Pachyderm `Repo` without making it part of a `Commit`. That means you have to start the `Commit`, write your data, then finish the `Commit`.

Here's an example:

```shell
$pachctl create-repo foo
$pachctl start-commit foo master
a6efb9ae6fba9e6fb9ae6
$echo 'hai' | pachctl put-file foo a6efb9ae6fba9e6fb9ae6 test.txt
$pachctl finish-commit foo a6efb9ae6fba9e6fb9ae6
# And writing in a new commit
$pachctl start-commit foo master
b2fecb2fe32bfec4bfced
$echo 'bai' | pachctl put-file foo b2fecb2fe32bfec4bfced test.txt
$pachctl finish-commit foo b2fecb2fe32bfec4bfced
```

In this example, we've written two words to the same file across two commits. 

You'll see that writing requires a `Commit` ID. If the `Commit` has been finished, you will only be able to read.

### Reading

Let's try reading the file we wrote above. That would look like this:

```shell
$pachctl get-file foo b2fecb2fe32bfec4bfceyd test.txt
hai
bai
```

Notice how the output is the cumulative result of the commits.

## Mounting PFS

Pachyderm uses FUSE to mount PFS. You can think of it simplistically as a network mount of PFS. While the files are truly served from object storage, you can see them locally by mounting PFS.

### Locally

Mounting PFS locally is a great way to debug an issue, or poke around PFS to understand how it works.

To mount locally, run:

```shell
$ mkdir ~/pfs
$ pachctl mount ~/pfs &
```

(If `~/pfs` already exists, you may need to `umount` it first)

Now you can look around the local mount:

```shell
# This is equivalent to `pachctl list-repo`
$ls ~/pfs
foo
# This is equiavelent to `pachctl list-commit foo`
$ls ~/pfs/foo
a6efb9ae6fba9e6fb9ae6
b2fecb2fe32bfec4bfced
# This is equiavelent to a call to `pachctl get-file ...`
$cat ~/pfs/foo/b2fecb2fe32bfec4bfced/test.txt
# And this has no CLI equivalent. It allows you to see all files in a commit:
$ls ~/pfs/foo/b2fecb2fe32bfec4bfced/
test.txt
```

Using this interface, you can grep, touch, ls, etc the files as you normally would. The exceptions are that you cannot write data to a commit that is finished and you cannot read data from a commit currently in progress.

