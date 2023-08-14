import {
  mockReposWithCommitQuery,
  mockReposQuery,
  Repo,
  OriginKind,
  mockRepoQuery,
  mockRepoWithLinkedPipelineQuery,
} from '@graphqlTypes';
import merge from 'lodash/merge';

import {buildPipeline} from './pipelines';

export const buildRepo = (repo: Partial<Repo>): Repo => {
  const defaultRepo = {
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
    linkedPipeline: null,
    lastCommit: null,
    authInfo: null,
    __typename: 'Repo',
  };

  return merge(defaultRepo, repo);
};

const REPO_MONTAGE: Repo = buildRepo({
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
});

const REPO_EDGES: Repo = buildRepo({
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
});

const REPO_IMAGES: Repo = buildRepo({
  branches: [
    {
      name: 'master',
      __typename: 'Branch',
    },
  ],
  createdAt: 1690221504,
  description: 'repo of images',
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
});

const REPO_EMPTY: Repo = buildRepo({
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
});
export const ALL_REPOS_WITH_COMMIT: Repo[] = [
  buildRepo({
    ...REPO_MONTAGE,
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
  }),
  buildRepo({
    ...REPO_EDGES,
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
  }),
  buildRepo({
    ...REPO_IMAGES,
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
  }),
  REPO_EMPTY,
];

export const ALL_REPOS: Repo[] = [
  REPO_MONTAGE,
  REPO_EDGES,
  REPO_IMAGES,
  REPO_EMPTY,
];

export const mockReposWithCommit = () =>
  mockReposWithCommitQuery((req, res, ctx) => {
    if (req.variables.args.projectId === 'default') {
      return res(ctx.data({repos: ALL_REPOS_WITH_COMMIT}));
    }
    return res();
  });

export const mockRepos = () =>
  mockReposQuery((req, res, ctx) => {
    if (req.variables.args.projectId === 'default') {
      return res(ctx.data({repos: ALL_REPOS}));
    }
    return res();
  });

export const mockRepoImages = () =>
  mockRepoQuery((req, res, ctx) => {
    const {projectId, id} = req.variables.args;
    if (projectId === 'default' && id === 'images') {
      return res(ctx.data({repo: REPO_IMAGES}));
    }
    return res();
  });

export const mockRepoMontage = () =>
  mockRepoQuery((req, res, ctx) => {
    const {projectId, id} = req.variables.args;
    if (projectId === 'default' && id === 'montage') {
      return res(ctx.data({repo: REPO_MONTAGE}));
    }
    return res();
  });

export const mockRepoImagesWithLinkedPipeline = () =>
  mockRepoWithLinkedPipelineQuery((req, res, ctx) => {
    const {projectId, id} = req.variables.args;
    if (projectId === 'default' && id === 'images') {
      return res(ctx.data({repo: REPO_IMAGES}));
    }
    return res();
  });

export const mockRepoWithLinkedPipeline = () =>
  mockRepoWithLinkedPipelineQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        repo: buildRepo({
          linkedPipeline: buildPipeline({}),
        }),
      }),
    );
  });

export const mockRepoWithNullLinkedPipeline = () =>
  mockRepoWithLinkedPipelineQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        repo: REPO_IMAGES,
      }),
    );
  });
