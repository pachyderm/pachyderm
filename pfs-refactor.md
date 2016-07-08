## Schema

`table` means a database table.
`object` means a JSON object.
`set<string>` can be simulated with map<string, bool> in JSON.

```
object DirAppend {
  set<string> children;
  bool delete;
}

object FileAppend {
  arr<string> blockrefs;
  bool delete;
}

object Branch {
  string name;
  int number;
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
  arr<DirAppend> dir_appends;
  arr<FileAppend> file_appends;
  int size;
  FileType file_type;
}

table Commit {
  string ID;  // primary key; UUID
  string repo;
  arr<Branch> branches;
  Timestamp started;
  Timestamp finished;
  arr<string> provenance;  // commit IDs, topologically sorted
}
```

## Code

`Table.Put` means putting a document into the table.

### StartCommit(repoName, parentCommitID = null, branch = null)

if parentCommit is null:
  if branch is null:
    branch = uuid()
  Branch.put(repoName+branch, repoName, branch)
  commit = newCommit(repoName)
  commit.branches += Branch(branch, 0)
else:
  parentCommit = getCommit(parentCommitID)
  commit = newCommit()
  commit.branches = parentCommit.branches
  if branch is null:
    commit.branches.lastElement.number += 1
  else:
    Branch.put(repoName+branch, repoName, branch)
    commit.branches += Branch(branch, 0)
return commit

### FinishCommit(commitID)

commit.finished = Now()

### MergeCommit(mergedID, commitIDs)

mergedCommit = getCommit(mergedID)
for commitID in commitIDs:
  diffs = getDiffs(commitID)
  for diff in diffs:
    // invariantHolds checks the invariant that there isn't a naming
    // conflict between files and directories.
    if !invariantHolds(diff):
      return "merge conflict"
    newDiff = clone(diff)
    newDiff.commitID = mergedID
    Diff.Put(newDiff)

### GetHistory(commitID, fromCommitID = null)

commit := getCommit(commitID)
query = null
for each prefix of commit.branches:
  from = null
  to = null
  for each branch in prefix:
    from += "branch.name, 0"
    to += "branch.name, branch.number"
  query += between(from, to)
return query(branchesIndex)

### InspectFile(commitID, path, fromCommitID = null)

history = GetHistory(commitID, fromCommitID)
for commit in history:
  query += getDiff(commit, path)
diffs = query()
return coalesce(diffs)

### GetFile(commitID, path, fromCommitID)

the same as InspectFile

### PutFile(commitID, path, reader)

blockrefs = read(reader)
Diff.Put(newDiff(commitID, path, blockrefs))

### CreateJob(parentCommit)

for shard in shards:
  shardCommits += StartCommit(repoName, parentCommit, branch = uuid())
for commit in shardCommits:
  runJob(commit)
outputCommit = StartCommit(repoName, parentCommit)
MergeCommit(outputCommit, shardCommits)
FinishCommit(outputCommit)

