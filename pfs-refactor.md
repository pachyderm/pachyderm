## Schema

### Notation

* `table` means a database table.
* `object` means a JSON object.
* `set<string>` can be simulated with map<string, bool> in JSON.
* `typeA | typeB` means either `typeA` or `typeB`.

### Schema

```
object Append {
  set<string> children;
  arr<string> blockrefs;
  bool delete;
}

object Clock {
  int branch;
  int commit;
}

table Repo {
  string name;  // primary key
  int branch_ID;  // monotonically increasing
  Timestamp created;
}

table Branch {
  string ID; // repo + name
  int branch_ID;
  string repo;
  string name; // a human readable name
}

table Diff {
  string ID;  // primary key; commitID + path
  string commitID;
  string path;
  arr<Clock> clocks;
  arr<Append> appends;
  int size;
}

table Commit {
  string ID;  // primary key; UUID
  string repo;
  arr<Clock> clocks;
  Timestamp started;
  Timestamp finished;
  arr<string> provenance;  // commit IDs, topologically sorted
}
```

## Code

### Notation

The code is all pseudocode.

* `Table.Operation` means a database operation on a table.
* `causedBy` is the vector clock comparison function.

### StartCommit(repoName, parentCommitID = null, branchName = null)

```
if parentCommit is null:
  if branchName is null:
    branchName = uuid()
  BranchTable.CreateIfNotExist(repoName+branchName, repoName, branch)
  commit = CommitTable.New()
  repo = RepoTable.Get(repoName)
  repo.BranchCounter += 1
  commit.Clocks += BranchClock(repo.BranchCounter, 0)
else:
  parentCommit = CommitTable.Get(parentCommitID)
  commit = parentCommit.Clone()
  if branch is null:
    commit.Clocks.lastElement.Clock += 1
  else:
    BranchTable.CreateIfNotExist(repoName+branchName, repoName, branch)
    repo = RepoTable.Get(repoName)
    repo.BranchCounter += 1
    commit.Clocks += BranchClock(repo.BranchCounter, 0)
return commit
```

### History(toCommit, fromCommit)

```
commitIndex = commitTable.CreateIndex(repo+vector)

History(toCommit, fromCommit):
    clocksTo = CommitTable.Get(toCommit).Clocks
    clocksFrom = CommitTable.Get(fromCommit).Clocks
    return Between(clocksTo, clocksFrom)  // this can be done mathematically
```

### MergeCommits(repoName, fromCommits, parentCommitId = null, branch = null, squash = false)

### InspectFile(path, toCommit, fromCommit = null)

```
diffIndex = diffTable.CreateIndex(repo+delete+path+vector)

InspectFile(path, toCommit, fromCommit):
    clocksTo = CommitTable.Get(toCommit).Clocks
    clocksFrom = CommitTable.Get(fromCommit).Clocks
    newClocksFrom = diffIndex.Between(repo+true+path+clocksFrom, repo+true+path+clocksTo).filter(caused_by).first().Clocks
    return Between(clocksTo, newClocksFrom)  // start streaming blocks as soon as we get the first diff; can be done mathematically
```

### GetFile(commitID, path, fromCommitID)

same thing as InspectFile

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
MergeCommits(repo, fromCommits, parentOutputCommit, squash = true)
```
