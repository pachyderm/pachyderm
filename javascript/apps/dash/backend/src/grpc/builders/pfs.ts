import {
  Branch,
  Commit,
  CommitInfo,
  File,
  FileInfo,
  FileType,
  Repo,
  Trigger,
} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {
  timestampFromObject,
  TimestampObject,
} from '@dash-backend/grpc/builders/protobuf';

export type FileObject = {
  commitId?: Commit.AsObject['id'];
  path?: File.AsObject['path'];
  branch?: BranchObject;
};

export type FileInfoObject = {
  committed: FileInfo.AsObject['committed'];
  file: FileObject;
  fileType: FileType;
  hash: FileInfo.Details.AsObject['hash'];
  sizeBytes: FileInfo.Details.AsObject['sizeBytes'];
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
  finished?: TimestampObject;
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

  const details = new FileInfo.Details();

  details.setHash(hash);
  details.setSizeBytes(sizeBytes);

  fileInfo.setDetails(details);

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

export const commitInfoFromObject = ({
  commit,
  description = '',
  sizeBytes = 0,
  started,
  finished,
}: CommitInfoObject) =>
  new CommitInfo()
    .setCommit(commitFromObject(commit))
    .setDescription(description)
    .setDetails(new CommitInfo.Details().setSizeBytes(sizeBytes))
    .setStarted(started ? timestampFromObject(started) : undefined)
    .setFinished(finished ? timestampFromObject(finished) : undefined);
