import merge from 'lodash/merge';
import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  ListRepoRequest,
  InspectRepoRequest,
  RepoInfo,
} from '@dash-frontend/api/pfs';
import {RequestError} from '@dash-frontend/api/utils/error';

export const buildRepo = (repo: Partial<RepoInfo> = {}): RepoInfo => {
  const defaultRepo: RepoInfo = {
    branches: [
      {
        name: 'master',
        __typename: 'Branch',
      },
    ],
    created: '2023-07-24T17:58:24.000Z',
    description: '',
    repo: {
      name: 'images',
      project: {
        name: 'default',
      },
    },
    sizeBytesUpperBound: '14783',
    authInfo: {
      roles: ['clusterAdmin'],
      __typename: 'AuthInfo',
    },
    __typename: 'RepoInfo',
  };

  return merge(defaultRepo, repo);
};

const REPO_INFO_MONTAGE: RepoInfo = buildRepo({
  created: '2023-07-24T17:58:26Z',
  description: 'Output repo for pipeline default/montage.',
  repo: {name: 'montage', project: {name: 'default'}},
  sizeBytesUpperBound: '31783',
});

const REPO_INFO_EDGES: RepoInfo = buildRepo({
  branches: [
    {
      name: 'master',
      __typename: 'Branch',
    },
  ],
  created: '2023-07-24T17:58:25Z',
  description: 'Output repo for pipeline default/edges.',
  repo: {name: 'edges', project: {name: 'default'}},
  sizeBytesUpperBound: '22783',
});

const REPO_INFO_IMAGES: RepoInfo = buildRepo({
  branches: [
    {
      name: 'master',
      __typename: 'Branch',
    },
  ],
  created: '2023-07-24T17:58:24Z',
  description: 'repo of images',
  repo: {name: 'images', project: {name: 'default'}},
  sizeBytesUpperBound: '14783',
});

export const REPO_INFO_EMPTY: RepoInfo = buildRepo({
  branches: [],
  created: '2023-07-24T17:58:14Z',
  description: '',
  repo: {name: 'empty', project: {name: 'default'}},
  sizeBytesUpperBound: '0',
  authInfo: {
    roles: ['clusterAdmin', 'repoOwner'],
    __typename: 'AuthInfo',
  },
});

export const ALL_REPO_INFOS: RepoInfo[] = [
  REPO_INFO_MONTAGE,
  REPO_INFO_EDGES,
  REPO_INFO_IMAGES,
  REPO_INFO_EMPTY,
];

export const mockRepos = () =>
  rest.post<ListRepoRequest, Empty, RepoInfo[]>(
    '/api/pfs_v2.API/ListRepo',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body['projects'][0]['name'] === 'default') {
        return res(ctx.json(ALL_REPO_INFOS));
      }
      return res(ctx.json([]));
    },
  );

export const mockRepoImages = () =>
  rest.post<InspectRepoRequest, Empty, RepoInfo>(
    '/api/pfs_v2.API/InspectRepo',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body.repo.name === 'images' && body.repo.project.name === 'default') {
        return res(ctx.json(REPO_INFO_IMAGES));
      }
      return res(ctx.json({}));
    },
  );

export const mockRepoEdges = () =>
  rest.post<InspectRepoRequest, Empty, RepoInfo | RequestError>(
    '/api/pfs_v2.API/InspectRepo',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body.repo.name === 'edges' && body.repo.project.name === 'default') {
        return res(ctx.json(REPO_INFO_EDGES));
      }
      return res(
        ctx.status(403),
        ctx.json({
          code: 5,
          message: 'not found',
          details: [],
        }),
      );
    },
  );

export const mockRepoMontage = () =>
  rest.post<InspectRepoRequest, Empty, RepoInfo>(
    '/api/pfs_v2.API/InspectRepo',
    async (req, res, ctx) => {
      const body = await req.json();
      if (
        body.repo.name === 'montage' &&
        body.repo.project.name === 'default'
      ) {
        return res(ctx.json(REPO_INFO_MONTAGE));
      }
      return res(ctx.json({}));
    },
  );

export const mockReposEmpty = () =>
  rest.post<ListRepoRequest, Empty, RepoInfo[]>(
    '/api/pfs_v2.API/ListRepo',
    (_req, res, ctx) => res(ctx.json([])),
  );

export const generatePagingRepos = (n: number): RepoInfo[] => {
  const repos: RepoInfo[] = [];
  for (let i = 0; i < n; i++) {
    repos.push(
      buildRepo({
        repo: {name: `repo-${i}`, project: {name: 'default'}},
      }),
    );
  }
  return repos;
};
