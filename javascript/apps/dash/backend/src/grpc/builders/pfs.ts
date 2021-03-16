import {
  Commit,
  File,
  FileInfo,
  FileType,
  Repo,
  Trigger,
} from '@pachyderm/proto/pb/pfs/pfs_pb';

export type FileObject = {
  commitId?: Commit.AsObject['id'];
  path?: File.AsObject['path'];
  repoName: Repo.AsObject['name'];
};

export type FileInfoObject = {
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

export type CommitObject = {
  repo?: RepoObject;
  id: Commit.AsObject['id'];
};

export const fileFromObject = ({
  commitId = 'master',
  path = '/',
  repoName,
}: FileObject) => {
  const file = new File();
  const commit = new Commit();
  const repo = new Repo();

  repo.setName(repoName);
  commit.setId(commitId);
  commit.setRepo(repo);
  file.setPath(path);
  file.setCommit(commit);

  return file;
};

export const fileInfoFromObject = ({
  file,
  fileType,
  hash,
  sizeBytes,
}: FileInfoObject) => {
  const fileInfo = new FileInfo();

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

  return repo;
};

export const commitFromObject = ({repo, id}: CommitObject) => {
  const commit = new Commit();

  if (repo) {
    commit.setRepo(repoFromObject(repo));
  }
  commit.setId(id);

  return commit;
};
