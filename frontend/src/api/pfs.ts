import {
  InspectProjectRequest,
  CreateProjectRequest,
  DeleteProjectRequest,
  CreateRepoRequest,
  ListRepoRequest,
  InspectRepoRequest,
  DeleteRepoRequest,
  InspectCommitRequest,
  DiffFileRequest,
  ProjectInfo,
  RepoInfo,
  CommitInfo,
  Commit,
  ListCommitRequest,
  ListBranchRequest,
  BranchInfo,
  DiffFileResponse,
  FindCommitsRequest,
  ListFileRequest,
  FileInfo,
  StartCommitRequest,
  FinishCommitRequest,
  GlobFileRequest,
} from '@dash-frontend/generated/proto/pfs/pfs.pb';

import {pps, pfs} from './base/api';
import {isAbortSignalError, RequestError} from './utils/error';
import {getHeaders, getRequestOptions} from './utils/requestHeaders';

export const inspectProject = async (req: InspectProjectRequest) => {
  return await pfs.InspectProject(req, getRequestOptions());
};

export const listProject = async () => {
  const projects: ProjectInfo[] = [];

  await pfs.ListProject(
    {},
    (project) => projects.push(project),
    getRequestOptions(),
  );

  return projects;
};

export const listRepo = async (req: ListRepoRequest) => {
  const repos: RepoInfo[] = [];

  await pfs.ListRepo(req, (repo) => repos.push(repo), getRequestOptions());

  return repos;
};

export const inspectRepo = async (req: InspectRepoRequest) => {
  return await pfs.InspectRepo(
    {repo: {...req.repo, type: 'user'}},
    getRequestOptions(),
  );
};

export const createRepo = async (req: CreateRepoRequest) => {
  return await pfs.CreateRepo(req, getRequestOptions());
};

export const deleteRepo = async (req: DeleteRepoRequest) => {
  return await pfs.DeleteRepo(
    {repo: {...req.repo, type: 'user'}},
    getRequestOptions(),
  );
};

export const listCommit = async (req: ListCommitRequest) => {
  const commits: CommitInfo[] = [];

  await pfs.ListCommit(
    req,
    (commit) => commits.push(commit),
    getRequestOptions(),
  );

  return commits;
};

export const listBranch = async (req: ListBranchRequest) => {
  const branches: BranchInfo[] = [];

  await pfs.ListBranch(
    req,
    (branch) => branches.push(branch),
    getRequestOptions(),
  );

  return branches;
};

export const listFiles = async (req: ListFileRequest) => {
  const files: FileInfo[] = [];

  await pfs.ListFile(req, (file) => files.push(file), getRequestOptions());

  return files;
};

export const inspectCommit = async (req: InspectCommitRequest) => {
  return await pfs.InspectCommit(req, getRequestOptions());
};

export const diffFile = async (req: DiffFileRequest) => {
  const fileDiffs: DiffFileResponse[] = [];

  await pfs.DiffFile(req, (diff) => fileDiffs.push(diff), getRequestOptions());

  return fileDiffs;
};

export const createProject = async (req: CreateProjectRequest) => {
  await pfs.CreateProject(req, getRequestOptions());
};

export const updateProject = async (req: CreateProjectRequest) => {
  await pfs.CreateProject({...req, update: true}, getRequestOptions());
};

export const deleteProject = async (req: DeleteProjectRequest) => {
  if (!req.project) {
    throw new Error('project required');
  }

  await pps.DeletePipelines(
    {projects: [req.project], force: req.force},
    getRequestOptions(),
  );
  await pfs.DeleteRepos(
    {projects: [req.project], force: req.force},
    getRequestOptions(),
  );
  await pfs.DeleteProject(req, getRequestOptions());
};

// Add any types or enums that don't exist in Pachyderm
export enum ProjectStatus {
  HEALTHY = 'HEALTHY',
  UNHEALTHY = 'UNHEALTHY',
}

export const listCommitsPaged = async (args: ListCommitRequest) => {
  const modifiedArgs = {...args};
  if (args?.number) modifiedArgs.number = String(Number(args.number) + 1);

  const commits = await listCommit(modifiedArgs);

  let nextCursor = undefined;

  // If commits.length is not greater than limit there are no pages left
  if (args?.number && commits && commits?.length > Number(args?.number)) {
    commits.pop(); // remove the extra commit from the response
    nextCursor = commits[commits.length - 1];
  }

  return {
    commits,
    cursor: nextCursor?.started,
    parentCommit: nextCursor?.parentCommit?.id,
  };
};

export const findCommitsPaged = async (req: FindCommitsRequest) => {
  const foundCommits: Commit[] = [];
  let lastSearchedCommit: Commit | undefined;

  await pfs.FindCommits(
    req,
    (commit) => {
      if (commit.foundCommit) {
        foundCommits.push(commit.foundCommit);
      } else {
        lastSearchedCommit = commit.lastSearchedCommit;
      }
    },
    getRequestOptions(),
  );

  const commits = await Promise.all(
    foundCommits.map(async (commit) => {
      return await pfs.InspectCommit({commit}, getRequestOptions());
    }),
  );

  let cursor = '';
  if (lastSearchedCommit) {
    const lastCommitInfo = await pfs.InspectCommit(
      {commit: lastSearchedCommit},
      getRequestOptions(),
    );
    cursor = lastCommitInfo.parentCommit?.id || '';
  }

  return {commits, cursor};
};

export const listFilesPaged = async (args: ListFileRequest) => {
  const originalNumber = Number(args?.number);
  if (args?.number) args.number = String(Number(args.number) + 1);

  const files = await listFiles(args);

  let nextCursor = undefined;

  // If files.length is not greater than limit there are no pages left
  if (args && args.number && files && files?.length > originalNumber) {
    files.pop(); // remove the extra file from the response
    nextCursor = files[files.length - 1];
  }

  return {
    files,
    cursor: nextCursor,
  };
};

export const globFile = async (req: GlobFileRequest, limit: number) => {
  const files: FileInfo[] = [];
  const controller = new AbortController();

  try {
    await pfs.GlobFile(
      req,
      (file) => {
        if (files.length < limit) {
          files.push(file);
        } else {
          controller.abort();
        }
      },
      {...getRequestOptions(), signal: controller.signal},
    );
  } catch (error) {
    if (isAbortSignalError(error)) {
      return files;
    }

    throw error;
  }

  return files;
};

export type EncodeArchiveUrlRequest = {
  projectId: string;
  repoId: string;
  commitId: string;
  paths: string[];
};
export type EncodeArchiveUrlResponse = {
  url: string;
};

export const encodeArchiveUrl = async (args: EncodeArchiveUrlRequest) => {
  return (await fetch('/encode/archive', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(args),
  }).then((res) => res.json())) as EncodeArchiveUrlResponse & RequestError;
};

export const startCommit = async (req: StartCommitRequest) => {
  return await pfs.StartCommit(req, getRequestOptions());
};

export const finishCommit = async (req: FinishCommitRequest) => {
  return await pfs.FinishCommit(req, getRequestOptions());
};

export const modifyFile = async (body: string) => {
  return await fetch('/api/pfs_v2.API/ModifyFile', {
    method: 'POST',
    headers: getHeaders(),
    body,
  });
};

// Export all of the pfs types
export * from '@dash-frontend/generated/proto/pfs/pfs.pb';
