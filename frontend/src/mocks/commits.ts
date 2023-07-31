import {mockGetCommitsQuery, Commit, OriginKind} from '@graphqlTypes';

import generateId from '@dash-frontend/lib/generateId';

export const IMAGE_COMMITS: Commit[] = [
  {
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
  },
  {
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
  },
  {
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
  },
];

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
