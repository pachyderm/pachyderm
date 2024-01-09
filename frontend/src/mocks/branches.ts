import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {BranchInfo, ListBranchRequest} from '@dash-frontend/api/pfs';

const BRANCH_MASTER: BranchInfo = {
  branch: {
    name: 'master',
    repo: {
      name: 'images',
      type: undefined,
      __typename: 'Repo',
    },
    __typename: 'Branch',
  },
  __typename: 'BranchInfo',
};

const BRANCH_TEST: BranchInfo = {
  branch: {
    name: 'test',
    repo: {
      name: 'images',
      type: undefined,
      __typename: 'Repo',
    },
    __typename: 'Branch',
  },
  __typename: 'BranchInfo',
};

export const mockGetBranchesMasterOnly = () =>
  rest.post<ListBranchRequest, Empty, BranchInfo[]>(
    '/api/pfs_v2.API/ListBranch',
    (_req, res, ctx) => {
      return res(ctx.json([BRANCH_MASTER]));
    },
  );

export const mockGetBranches = () =>
  rest.post<ListBranchRequest, Empty, BranchInfo[]>(
    '/api/pfs_v2.API/ListBranch',
    (_req, res, ctx) => {
      return res(ctx.json([BRANCH_MASTER, BRANCH_TEST]));
    },
  );
