import {
  Branch,
  Commit,
  CommitInfo,
  CreateRepoRequest,
  DeleteRepoRequest,
  CommitSet,
  InspectCommitSetRequest,
  File,
  FileInfo,
  FileType,
  CreateBranchRequest,
  ListBranchRequest,
  DeleteBranchRequest,
  Repo,
  Trigger,
  StartCommitRequest,
  FinishCommitRequest,
  ListCommitRequest,
  SubscribeCommitRequest,
  InspectCommitRequest,
  CommitState,
  OriginKind,
} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {timestampFromObject, TimestampObject} from '../builders/protobuf';

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

export type StartCommitRequestObject = {
  branch: BranchObject;
  description?: StartCommitRequest.AsObject['description'];
  parent?: CommitObject;
};

export type FinishCommitRequestObject = {
  commit: CommitObject;
  error?: FinishCommitRequest.AsObject['error'];
  force?: FinishCommitRequest.AsObject['force'];
  description?: FinishCommitRequest.AsObject['description'];
};

export type InspectCommitRequestObject = {
  wait: CommitState;
  commit: CommitObject;
};

export type ListCommitRequestObject = {
  repo: RepoObject;
  number?: ListCommitRequest.AsObject['number'];
  reverse?: ListCommitRequest.AsObject['reverse'];
  all?: ListCommitRequest.AsObject['all'];
  originKind?: OriginKind;
  from?: CommitObject;
  to?: CommitObject;
};

export type SubscribeCommitRequestObject = {
  repo: RepoObject;
  branch?: SubscribeCommitRequest.AsObject['branch'];
  state?: CommitState;
  all?: SubscribeCommitRequest.AsObject['all'];
  originKind?: OriginKind;
  from?: CommitObject;
};

export type CreateBranchRequestObject = {
  head?: CommitObject;
  branch?: BranchObject;
  provenance: BranchObject[];
  trigger?: TriggerObject;
  newCommitSet: CreateBranchRequest.AsObject['newCommitSet'];
};

export type ListBranchRequestObject = {
  repo: RepoObject;
  reverse?: ListBranchRequest.AsObject['reverse'];
};

export type DeleteBranchRequestObject = {
  branch: BranchObject;
  force?: DeleteBranchRequest.AsObject['force'];
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
};

export type CommitSetObject = {
  id: CommitSet.AsObject['id'];
};

export type InspectCommitSetRequestObject = {
  commitSet: CommitSetObject;
  wait?: InspectCommitSetRequest.AsObject['wait'];
};

export type CreateRepoRequestObject = {
  repo: RepoObject;
  description?: CreateRepoRequest.AsObject['description'];
  update?: CreateRepoRequest.AsObject['update'];
};

export type DeleteRepoRequestObject = {
  repo: RepoObject;
  force?: DeleteRepoRequest.AsObject['force'];
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

export const startCommitRequestFromObject = ({
  branch,
  parent,
  description = '',
}: StartCommitRequestObject) => {
  const request = new StartCommitRequest();

  request.setBranch(branchFromObject(branch));
  if (parent) {
    request.setParent(commitFromObject(parent));
  }
  request.setDescription(description);

  return request;
};

export const finishCommitRequestFromObject = ({
  error,
  force = false,
  commit,
  description = '',
}: FinishCommitRequestObject) => {
  const request = new FinishCommitRequest();

  request.setForce(force);
  if (error) {
    request.setError(error);
  }
  if (commit) {
    request.setCommit(commitFromObject(commit));
  }
  request.setDescription(description);

  return request;
};

export const inspectCommitRequestFromObject = ({
  wait,
  commit,
}: InspectCommitRequestObject) => {
  const request = new InspectCommitRequest();

  if (wait) {
    request.setWait(wait);
  }

  if (commit) {
    request.setCommit(commitFromObject(commit));
  }

  return request;
};

export const listCommitRequestFromObject = ({
  number,
  all = true,
  originKind,
  from,
  to,
  repo,
  reverse = false,
}: ListCommitRequestObject) => {
  const request = new ListCommitRequest();

  if (repo) {
    request.setRepo(repoFromObject(repo).setType('user'));
  }

  if (from) {
    request.setFrom(commitFromObject(from));
  }

  if (to) {
    request.setTo(commitFromObject(to));
  }

  if (number) {
    request.setNumber(number);
  }

  if (originKind) {
    request.setOriginKind(originKind);
  }

  request.setAll(all);
  request.setReverse(reverse);

  return request;
};

export const subscribeCommitRequestFromObject = ({
  repo,
  branch,
  state,
  all = true,
  originKind,
  from,
}: SubscribeCommitRequestObject) => {
  const request = new SubscribeCommitRequest();

  request.setRepo(repoFromObject(repo).setType('user'));

  if (from) {
    request.setFrom(commitFromObject(from));
  }

  if (branch) {
    request.setBranch(branch);
  }

  if (state) {
    request.setState(state);
  }

  if (originKind) {
    request.setOriginKind(originKind);
  }

  request.setAll(all);

  return request;
};

export const createBranchRequestFromObject = ({
  head,
  branch,
  trigger,
  provenance,
  newCommitSet,
}: CreateBranchRequestObject) => {
  const request = new CreateBranchRequest();

  if (head) request.setHead(commitFromObject(head));
  if (branch) request.setBranch(branchFromObject(branch));

  if (provenance) {
    const provenanceArray: Branch[] = provenance.map((eachProvenanceObject) => {
      return branchFromObject(eachProvenanceObject);
    });
    request.setProvenanceList(provenanceArray);
  }

  if (trigger) request.setTrigger(triggerFromObject(trigger));
  request.setNewCommitSet(newCommitSet);

  return request;
};

export const listBranchRequestFromObject = ({
  repo,
  reverse = false,
}: ListBranchRequestObject) => {
  const request = new ListBranchRequest();

  request.setRepo(new Repo().setName(repo?.name || '').setType('user'));
  request.setReverse(reverse);

  return request;
};

export const deleteBranchRequestFromObject = ({
  branch,
  force = false,
}: DeleteBranchRequestObject) => {
  const request = new DeleteBranchRequest();

  if (branch) {
    request.setBranch(
      new Branch()
        .setName(branch.name)
        .setRepo(new Repo().setName(branch.repo?.name || '').setType('user')),
    );
  }

  request.setForce(force);

  return request;
};

export const commitInfoFromObject = ({
  commit,
  description = '',
  sizeBytes = 0,
  started,
  finishing,
  finished,
  sizeBytesUpperBound,
}: CommitInfoObject) => {
  const commitInfo = new CommitInfo()
    .setCommit(commitFromObject(commit))
    .setDescription(description)
    .setDetails(new CommitInfo.Details().setSizeBytes(sizeBytes))
    .setStarted(started ? timestampFromObject(started) : undefined)
    .setFinishing(finishing ? timestampFromObject(finishing) : undefined)
    .setFinished(finished ? timestampFromObject(finished) : undefined);
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

export const inspectCommitSetRequestFromObject = ({
  commitSet,
  wait = true,
}: InspectCommitSetRequestObject) => {
  const request = new InspectCommitSetRequest();

  request.setCommitSet(commitSetFromObject(commitSet));
  request.setWait(wait);

  return request;
};

export const createRepoRequestFromObject = ({
  repo,
  description = '',
  update = false,
}: CreateRepoRequestObject) => {
  const request = new CreateRepoRequest();

  request.setRepo(repoFromObject(repo));
  request.setDescription(description);
  request.setUpdate(update);

  return request;
};

export const deleteRepoRequestFromObject = ({
  repo,
  force = false,
}: DeleteRepoRequestObject) => {
  const request = new DeleteRepoRequest();

  request.setRepo(repoFromObject(repo));
  request.setForce(force);

  return request;
};
