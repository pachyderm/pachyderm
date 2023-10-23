import {
  mockGetCommitsQuery,
  Commit,
  OriginKind,
  mockCommitQuery,
  mockCommitDiffQuery,
} from '@graphqlTypes';

import generateId from '@dash-frontend/lib/generateId';

const COMMIT_4E: Commit = {
  repoName: 'images',
  branch: {
    name: 'master',
    repo: null,
    __typename: 'Branch',
  },
  description: 'commit not finished',
  originKind: OriginKind.USER,
  id: '4eb1aa567dab483f93a109db4641ee75',
  started: 1690221505,
  finished: -1,
  sizeBytes: 0,
  sizeDisplay: '0 B',
  __typename: 'Commit',
};

export const COMMIT_4A: Commit = {
  repoName: 'images',
  branch: {
    name: 'master',
    repo: null,
    __typename: 'Branch',
  },
  description: 'added mako',
  originKind: OriginKind.USER,
  id: '4a83c74809664f899261baccdb47cd90',
  started: 1690221505,
  finished: 1690321505,
  sizeBytes: 139232,
  sizeDisplay: '139.24 kB',
  __typename: 'Commit',
};

export const COMMIT_C4: Commit = {
  repoName: 'images',
  branch: {
    name: 'master',
    repo: null,
    __typename: 'Branch',
  },
  description: 'sold materia',
  originKind: OriginKind.USER,
  id: 'c43fffd650a24b40b7d9f1bf90fcfdbe',
  started: 1690221505,
  finished: 1690221505,
  sizeBytes: 58644,
  sizeDisplay: '58.65 kB',
  __typename: 'Commit',
};

export const IMAGE_COMMITS: Commit[] = [COMMIT_4E, COMMIT_4A, COMMIT_C4];

type generateCommitsArgs = {
  n: number;
  repoName?: string;
  branchName?: string;
};

export const generatePagingCommits = ({
  n,
  repoName = 'repo',
  branchName = 'master',
}: generateCommitsArgs) => {
  const commits: Commit[] = [];

  for (let i = 0; i < n; i++) {
    commits.push({
      repoName,
      branch: {
        name: branchName,
        repo: null,
        __typename: 'Branch',
      },
      description: `item ${i}`,
      originKind: OriginKind.USER,
      id: generateId(),
      started: i,
      finished: -1,
      sizeBytes: 0,
      sizeDisplay: '0 B',
      __typename: 'Commit',
    });
  }
  return commits;
};

export const mockGetImageCommits = () =>
  mockGetCommitsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        commits: {items: IMAGE_COMMITS, cursor: null, parentCommit: null},
      }),
    );
  });
export const mockGetCommitsA4Only = () =>
  mockGetCommitsQuery((req, res, ctx) => {
    const {projectId, repoName, branchName} = req.variables.args;
    if (
      projectId === 'default' &&
      repoName === 'images' &&
      branchName === 'master'
    ) {
      return res(
        ctx.data({
          commits: {items: [COMMIT_4A], cursor: null, parentCommit: null},
        }),
      );
    }
    return res();
  });

export const mockGetCommitA4 = () =>
  mockCommitQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        commit: COMMIT_4A,
      }),
    );
  });

export const mockGetCommitC4 = () =>
  mockCommitQuery((req, res, ctx) => {
    if (req.variables.args.id === 'c43fffd650a24b40b7d9f1bf90fcfdbe') {
      return res(
        ctx.data({
          commit: COMMIT_C4,
        }),
      );
    } else {
      return res(ctx.errors([]));
    }
  });

export const mockEmptyCommitDiff = () =>
  mockCommitDiffQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        commitDiff: null,
      }),
    );
  });

export const mockEmptyCommit = () =>
  mockCommitQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        commit: null,
      }),
    );
  });
