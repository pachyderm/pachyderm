import {
  Branch,
  Commit,
  CommitSet,
  Repo,
  DeleteFile,
  Project,
} from '../proto/pfs/pfs_pb';

export type RepoObject = {
  name: Repo.AsObject['name'];
  project?: Project.AsObject;
};

export type BranchObject = {
  name: Branch.AsObject['name'];
  repo?: RepoObject;
};

export type CommitObject = {
  id: Commit.AsObject['id'];
  branch?: BranchObject;
};

export type DeleteFileObject = {
  path: string;
};

export type CommitSetObject = {
  id: CommitSet.AsObject['id'];
};

export const commitFromObject = ({branch, id}: CommitObject) => {
  const commit = new Commit();

  if (branch) {
    commit.setBranch(
      new Branch()
        .setName(branch.name)
        .setRepo(new Repo().setName(branch.repo?.name || '').setType('user')),
    );
    if (branch.repo?.project?.name) {
      commit
        .getBranch()
        ?.getRepo()
        ?.setProject(new Project().setName(branch.repo?.project?.name));
    }
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

export const commitSetFromObject = ({id}: CommitSetObject) => {
  const commitSet = new CommitSet();

  commitSet.setId(id);

  return commitSet;
};

export const deleteFileFromObject = ({path}: DeleteFileObject) => {
  const deleteFile = new DeleteFile().setPath(path);
  return deleteFile;
};
