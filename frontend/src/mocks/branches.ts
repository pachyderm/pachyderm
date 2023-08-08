import {Branch, mockGetBranchesQuery} from '@graphqlTypes';

const BRANCH_MASTER: Branch = {
  name: 'master',
  repo: {
    name: 'images',
    type: null,
    __typename: 'RepoInfo',
  },
  __typename: 'Branch',
};

const BRANCH_TEST: Branch = {
  name: 'test',
  repo: {
    name: 'images',
    type: null,
    __typename: 'RepoInfo',
  },
  __typename: 'Branch',
};

export const mockGetBranchesMasterOnly = () =>
  mockGetBranchesQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        branches: [BRANCH_MASTER],
      }),
    );
  });

export const mockGetBranches = () =>
  mockGetBranchesQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        branches: [BRANCH_MASTER, BRANCH_TEST],
      }),
    );
  });
