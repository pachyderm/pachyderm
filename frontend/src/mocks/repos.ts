import {
  mockReposWithCommitQuery,
  mockReposQuery,
  Repo,
  OriginKind,
} from '@graphqlTypes';

export const ALL_REPOS_WITH_COMMIT: Repo[] = [
  {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    createdAt: 1690221506,
    description: 'Output repo for pipeline default/montage.',
    id: 'montage',
    name: 'montage',
    sizeDisplay: '0 B',
    sizeBytes: 31783,
    access: true,
    projectId: 'default',
    authInfo: {
      rolesList: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
    lastCommit: {
      repoName: 'montage',
      branch: {
        name: 'master',
        repo: null,
        __typename: 'Branch',
      },
      description: '',
      originKind: OriginKind.AUTO,
      id: '544e62c9a16f4a0aa995eb993ded1e52',
      started: 1690221518,
      finished: 1690221700,
      sizeBytes: 797,
      sizeDisplay: '797 B',
      __typename: 'Commit',
    },
  },
  {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    createdAt: 1690221505,
    description: 'Output repo for pipeline default/edges.',
    id: 'edges',
    name: 'edges',
    sizeDisplay: '22.79 kB',
    sizeBytes: 22783,
    access: true,
    projectId: 'default',
    authInfo: {
      rolesList: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
    lastCommit: {
      repoName: 'edges',
      branch: {
        name: 'master',
        repo: null,
        __typename: 'Branch',
      },
      description: '',
      originKind: OriginKind.USER,
      id: '544e62c9a16f4a0aa995eb993ded1e52',
      started: 1690221518,
      finished: 1690221518,
      sizeBytes: 1848,
      sizeDisplay: '1.85 kB',
      __typename: 'Commit',
    },
  },
  {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    createdAt: 1690221504,
    description: '',
    id: 'images',
    name: 'images',
    sizeDisplay: '0 B',
    sizeBytes: 14783,
    access: true,
    projectId: 'default',
    authInfo: {
      rolesList: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
    lastCommit: {
      repoName: 'images',
      branch: {
        name: 'master',
        repo: null,
        __typename: 'Branch',
      },
      description: '',
      originKind: OriginKind.USER,
      id: '4eb1aa567dab483f93a109db4641ee75',
      started: 1690221505,
      finished: -1,
      sizeBytes: 0,
      sizeDisplay: '0 B',
      __typename: 'Commit',
    },
  },
  {
    branches: [],
    createdAt: 1690221494,
    description: '',
    id: 'empty',
    name: 'empty',
    sizeDisplay: '0 B',
    sizeBytes: 0,
    access: true,
    projectId: 'ProjectA',
    authInfo: {
      rolesList: ['clusterAdmin', 'repoOwner'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
    lastCommit: null,
  },
];

export const ALL_REPOS: Repo[] = [
  {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    createdAt: 1690221506,
    description: 'Output repo for pipeline default/montage.',
    id: 'montage',
    name: 'montage',
    sizeDisplay: '0 B',
    sizeBytes: 31783,
    access: true,
    projectId: 'default',
    authInfo: {
      rolesList: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
  },
  {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    createdAt: 1690221505,
    description: 'Output repo for pipeline default/edges.',
    id: 'edges',
    name: 'edges',
    sizeDisplay: '22.79 kB',
    sizeBytes: 22783,
    access: true,
    projectId: 'default',
    authInfo: {
      rolesList: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
  },
  {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    createdAt: 1690221504,
    description: '',
    id: 'images',
    name: 'images',
    sizeDisplay: '0 B',
    sizeBytes: 14783,
    access: true,
    projectId: 'default',
    authInfo: {
      rolesList: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
  },
  {
    branches: [],
    createdAt: 1690221494,
    description: '',
    id: 'empty',
    name: 'empty',
    sizeDisplay: '0 B',
    sizeBytes: 0,
    access: true,
    projectId: 'ProjectA',
    authInfo: {
      rolesList: ['clusterAdmin', 'repoOwner'],
      __typename: 'AuthInfo',
    },
    __typename: 'Repo',
  },
];

export const mockReposWithCommit = () =>
  mockReposWithCommitQuery((_req, res, ctx) => {
    return res(ctx.data({repos: ALL_REPOS_WITH_COMMIT}));
  });

export const mockRepos = () =>
  mockReposQuery((_req, res, ctx) => {
    return res(ctx.data({repos: ALL_REPOS}));
  });
