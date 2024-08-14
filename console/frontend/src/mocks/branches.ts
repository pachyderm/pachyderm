import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {BranchInfo, ListBranchRequest} from '@dash-frontend/api/pfs';

const BRANCH_MASTER: BranchInfo = {
  branch: {
    name: 'master',
    repo: {
      name: 'images',
      type: undefined,
    },
  },
  head: {
    id: 'c43fffd650a24b40b7d9f1bf90fcfdbe',
  },
};

const BRANCH_NEW: BranchInfo = {
  branch: {
    name: 'new',
    repo: {
      name: 'images',
      type: undefined,
    },
  },
  head: {
    id: '4a83c74809664f899261baccdb47cd90',
  },
};

const BRANCH_SAMPLE: BranchInfo = {
  branch: {
    name: 'sample',
    repo: {
      name: 'images',
      type: undefined,
    },
  },
  head: {
    id: '4eb1aa567dab483f93a109db4641ee75',
  },
};

export const BRANCH_RENDER: BranchInfo = {
  branch: {
    name: 'render',
    repo: {
      name: 'images',
      type: undefined,
    },
  },
  head: {
    id: '252d1850a5fa484ca7320ce1091cf483',
  },
};

const BRANCH_RENDERTWO: BranchInfo = {
  branch: {
    name: 'renderTwo',
    repo: {
      name: 'images',
      type: undefined,
    },
  },
  head: {
    id: '252d1850a5fa484ca7320ce1091cf483',
  },
};

const BRANCH_TEST: BranchInfo = {
  branch: {
    name: 'test',
    repo: {
      name: 'images',
      type: undefined,
    },
  },
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
      return res(
        ctx.json([
          BRANCH_MASTER,
          BRANCH_SAMPLE,
          BRANCH_NEW,
          BRANCH_RENDER,
          BRANCH_RENDERTWO,
          BRANCH_TEST,
        ]),
      );
    },
  );
