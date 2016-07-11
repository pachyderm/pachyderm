## Schema

### Notation

* `table` means a database table.
* `object` means a JSON object.
* `set<string>` can be simulated with map<string, bool> in JSON.
* `typeA | typeB` means either `typeA` or `typeB`.
* `Table.Operation` means a database operation on a table.

### Schema

```
object DirAppend {
  set<string> children;
  bool delete;
}

object FileAppend {
  arr<string> blockrefs;
  bool delete;
}

object BranchClock {
  string name;
  int clock;
}

table Repo {
  string name;  // primary key
  Timestamp created;
}

table Branch {
  string ID;  // primary key; repo name + branch name
  string repo;
  string name;
}

table Diff {
  string ID;  // primary key; commitID + path
  string commitID;
  string path;
  // If file_type == dir
  FileType file_type;
  arr<DirAppend> dir_appends;
  arr<FileAppend> file_appends;
  int size;
}

table Commit {
  string ID;  // primary key; UUID
  string repo;
  arr<BranchClock | arr<BranchClock>> branch_vector;
  Timestamp started;
  Timestamp finished;
  arr<string> provenance;  // commit IDs, topologically sorted
}
```

## Code

### GetHistory(commitID, fromCommitID = null)

Given a commitID and a fromCommitID, `GetHistory` returns all commits between the two commits.

In our schema, each commit carries an immutable `branch_vector`, which is similar to a [vector clock](https://en.wikipedia.org/wiki/Vector_clock) of branches.  Concretely, we assign `branch_vector` to commits using the following rules:

1. If the commit has no parent, its `branch_vector` is `[(new_branch_name, 0)]`, where `new_branch_name` can be specified by the user, or can be just a UUID.

2. If the commit has a parent, its `branch_vector` is the same as its parent's `branch_vector`, with the last element incremented by 1.

    For example, if the parent is `[(x, 1), (y, 0), (z, 1)]`, the new commit is `[(x, 1), (y, 0), (z, 2)]`. 

3. If the commit is the result of merging multiple commits, whose `branch_vector`s are `V1, V2, ... Vn` respectively, then its branch vector is `[[V1, V2, ..., Vn], (new_branch_name, 0)]`.

    For example, if we are merging commit `[(x, 1), (y, 0)]` and `[(x, 1), (z, 1)]` , the new commit is `[[[(x, 1), (y, 0)], [(x, 1), (z, 1)]], (new_branch_name, 0)]`

Note that `branch_vector` is not strictly a vector clock itself.  Rather, it's a compact representation of many smaller `branch_vector`s.  We define `extract()` as a function that extracts `branch_vector`s from a `branch_vector`.  `extract` works as follows:

```
func extract(bv):
  result = []
  last = lastElement(bv)
  if last is not an array:
    result += bv
    bv.pop()  // remove the last element
  else:
    bv.pop()  // remove the last element
    for element in last:
      result += bv + extract(element)
```

Define E(X) as extract(bv) where bv is x's branch vector.  The following property holds:

    The ancestors of X are simply all the commits whose `branch_vector`s are smaller than at least one element in E(x).  

### StartCommit(repoName, parentCommitID = null, branch = null)

```
if parentCommit is null:
  if branch is null:
    branch = uuid()
  Branch.create_if_not_exist(repoName+branch, repoName, branch)
  commit = newCommit(repoName)
  commit.branch_vector += BranchClock(branch, 0)
else:
  parentCommit = getCommit(parentCommitID)
  commit = newCommit()
  commit.branch_vector = parentCommit.branch_vector
  if branch is null:
    commit.branch_vector.lastElement.clock += 1
  else:
    Branch.create_if_not_exist(repoName+branch, repoName, branch)
    commit.branch_vector += BranchClock(branch, 0)
return commit
```

### MergeCommits(repoName, fromCommits, parentCommitId = null, branch = null)

`MergeCommit` creates a new commit whose parents are `fromCommits`.  Note that the commits in `fromCommits` can be open or closed.

Here we only describe the merging of `branch_vector`s which is the only interesting part.

```
TODO
```

### InspectFile(commitID, path, fromCommitID = null)

```
history = GetHistory(commitID, fromCommitID)
for commit in history:
  query += getDiff(commit, path)
diffs = query()
return coalesce(diffs)
```

### GetFile(commitID, path, fromCommitID)

the same as InspectFile

### PutFile(commitID, path, reader)

```
blockrefs = read(reader)
diff = Diff.CreateIfNotExist(commitID, path)
diff.blockrefs.append(blockrefs)
```

### CreateJob

A job with a parallelism of N creates N branches based off the parent output commit.  Each shard operates on its own branch.  Once all shard finishes, we merge all branches into the original branch.

Note that since every shard operates on its own branch, there is no chance of blockrefs interleaving.  Therefore, we are able to do without "handles".

```
fromCommits = []
for shard in shards:
    commit = StartCommit(repo, parentOutputCommit)
    runJob(commit)
    fromCommits += commit
MergeCommits(repo, fromCommits, parentOutputCommit)
```
