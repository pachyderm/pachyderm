import {timestampFromObject, TimestampObject} from '../builders/protobuf';
import {
  Branch,
  Commit,
  CommitInfo,
  CommitSet,
  File,
  FileInfo,
  FileType,
  Repo,
  Trigger,
  OriginKind,
  CommitOrigin,
  DeleteFile,
} from '../proto/pfs/pfs_pb';

export type FileObject = {
  commitId?: Commit.AsObject['id'];
  path?: File.AsObject['path'];
  branch?: BranchObject;
};

export type FileInfoObject = {
  committed: FileInfo.AsObject['committed'];
  file: FileObject;
  fileType: FileType;
  hash: FileInfo.AsObject['hash'];
  sizeBytes: FileInfo.AsObject['sizeBytes'];
};

export type TriggerObject = {
  branch: Trigger.AsObject['branch'];
  all: Trigger.AsObject['all'];
  cronSpec: Trigger.AsObject['cronSpec'];
  size: Trigger.AsObject['size'];
  commits: Trigger.AsObject['commits'];
};

export type RepoObject = {
  name: Repo.AsObject['name'];
};

export type BranchObject = {
  name: Branch.AsObject['name'];
  repo?: RepoObject;
};

export type CommitObject = {
  id: Commit.AsObject['id'];
  branch?: BranchObject;
};

export type CommitInfoObject = {
  commit: CommitObject;
  description?: CommitInfo.AsObject['description'];
  sizeBytes?: CommitInfo.Details.AsObject['sizeBytes'];
  started?: TimestampObject;
  finishing?: TimestampObject;
  finished?: TimestampObject;
  sizeBytesUpperBound?: CommitInfo.AsObject['sizeBytesUpperBound'];
  originKind?: OriginKind;
};

export type DeleteFileObject = {
  path: string;
};

export type CommitSetObject = {
  id: CommitSet.AsObject['id'];
};

export const fileFromObject = ({
  commitId = 'master',
  path = '/',
  branch,
}: FileObject) => {
  const file = new File();
  const commit = new Commit();
  const repo = new Repo();
  repo.setType('user');
  let repoBranch = new Branch();

  commit.setId(commitId);

  if (branch) {
    repoBranch = new Branch().setName(branch.name);

    if (branch.repo) {
      repo.setName(branch.repo.name);
      repoBranch.setRepo(repo);
      commit.setBranch(repoBranch);
    }
  }

  file.setPath(path);
  file.setCommit(commit);

  return file;
};

export const fileInfoFromObject = ({
  committed,
  file,
  fileType,
  hash,
  sizeBytes,
}: FileInfoObject) => {
  const fileInfo = new FileInfo();

  if (committed) {
    fileInfo.setCommitted(timestampFromObject(committed));
  }

  fileInfo.setFile(fileFromObject(file));
  fileInfo.setFileType(fileType);
  fileInfo.setHash(hash);
  fileInfo.setSizeBytes(sizeBytes);

  return fileInfo;
};

export const triggerFromObject = ({
  branch,
  all,
  cronSpec,
  size,
  commits,
}: TriggerObject) => {
  const trigger = new Trigger();
  trigger.setBranch(branch);
  trigger.setAll(all);
  trigger.setCronSpec(cronSpec);
  trigger.setSize(size);
  trigger.setCommits(commits);

  return trigger;
};

export const repoFromObject = ({name}: RepoObject) => {
  const repo = new Repo();
  repo.setName(name);
  repo.setType('user');

  return repo;
};

export const commitFromObject = ({branch, id}: CommitObject) => {
  const commit = new Commit();

  if (branch) {
    commit.setBranch(
      new Branch()
        .setName(branch.name)
        .setRepo(new Repo().setName(branch.repo?.name || '').setType('user')),
    );
  }
  commit.setId(id);

  return commit;
};

export const branchFromObject = ({name, repo}: BranchObject) => {
  const branch = new Branch();
  branch.setName(name);
  branch.setRepo(new Repo().setName(repo?.name || '').setType('user'));

  return branch;
};

export const commitInfoFromObject = ({
  commit,
  description = '',
  sizeBytes = 0,
  started,
  finishing,
  finished,
  sizeBytesUpperBound,
  originKind,
}: CommitInfoObject) => {
  const commitInfo = new CommitInfo()
    .setCommit(commitFromObject(commit))
    .setDescription(description)
    .setDetails(new CommitInfo.Details().setSizeBytes(sizeBytes))
    .setStarted(started ? timestampFromObject(started) : undefined)
    .setFinishing(finishing ? timestampFromObject(finishing) : undefined)
    .setFinished(finished ? timestampFromObject(finished) : undefined)
    .setOrigin(new CommitOrigin().setKind(originKind || OriginKind.AUTO));
  if (sizeBytesUpperBound) {
    commitInfo.setSizeBytesUpperBound(sizeBytesUpperBound);
  }
  return commitInfo;
};
export const commitSetFromObject = ({id}: CommitSetObject) => {
  const commitSet = new CommitSet();

  commitSet.setId(id);

  return commitSet;
};

export const deleteFileFromObject = ({path}: DeleteFileObject) => {
  const deleteFile = new DeleteFile().setPath(path);
  return deleteFile;
};
