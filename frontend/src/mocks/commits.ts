import merge from 'lodash/merge';
import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  CommitInfo,
  OriginKind,
  ListCommitRequest,
  DiffFileRequest,
  DiffFileResponse,
  FileType,
  StartCommitRequest,
  Commit,
  FinishCommitRequest,
  InspectCommitRequest,
} from '@dash-frontend/api/pfs';
import {RequestError} from '@dash-frontend/api/utils/error';
import {getISOStringFromUnix} from '@dash-frontend/lib/dateTime';
import generateId from '@dash-frontend/lib/generateId';

export const buildCommit = (commit: CommitInfo) => {
  const defaultCommit: CommitInfo = {
    __typename: 'CommitInfo',
    origin: {kind: OriginKind.USER},
  };

  return merge(defaultCommit, commit);
};

export const COMMIT_INFO_4E: CommitInfo = buildCommit({
  commit: {
    id: '4eb1aa567dab483f93a109db4641ee75',
    repo: {name: 'images', project: {name: 'default'}},
    branch: {
      name: 'master',
      __typename: 'Branch',
    },
  },
  description: 'commit not finished',
  started: '2023-07-24T17:58:25Z',
  finished: undefined,
  sizeBytesUpperBound: '0',
  details: {
    sizeBytes: '0',
  },
});

export const COMMIT_INFO_4A: CommitInfo = buildCommit({
  commit: {
    id: '4a83c74809664f899261baccdb47cd90',
    repo: {name: 'images', project: {name: 'default'}},
    branch: {
      name: 'master',
      __typename: 'Branch',
    },
  },
  description: 'added mako',
  started: '2023-07-24T17:58:25Z',
  finished: '2023-07-25T21:45:05Z',
  sizeBytesUpperBound: '139232',
  details: {
    sizeBytes: '139232',
  },
  parentCommit: {id: '4eb1aa567dab483f93a109db4641ee75'},
});

export const COMMIT_INFO_C4: CommitInfo = buildCommit({
  commit: {
    id: 'c43fffd650a24b40b7d9f1bf90fcfdbe',
    repo: {name: 'images', project: {name: 'default'}},
    branch: {
      name: 'master',
      __typename: 'Branch',
    },
  },
  description: 'sold materia',
  started: '2023-07-24T17:58:25Z',
  finished: '2023-07-24T17:58:25Z',
  sizeBytesUpperBound: '58644',
  details: {
    sizeBytes: '58644',
  },
  parentCommit: {id: 'c43fffd650a24b40b7d9f1bf90fcfdbe'},
});

export const IMAGE_COMMITS: CommitInfo[] = [
  COMMIT_INFO_4A,
  COMMIT_INFO_4E,
  COMMIT_INFO_C4,
];

export const COMMIT_INFO_G2_NO_BRANCH: CommitInfo = buildCommit({
  commit: {
    id: 'g2bb3e50cd124b76840145a8c18f8892',
    repo: {name: 'images', project: {name: 'default'}},
  },
  description: 'I deleted this branch',
  started: '2023-07-24T17:58:25Z',
  finished: '2023-07-24T17:58:25Z',
  sizeBytesUpperBound: '139232',
  details: {
    sizeBytes: '139232',
  },
  parentCommit: {id: '73fe17002aab4126a671c042678a62b2'},
});

export const COMMIT_INFO_73_NO_BRANCH: CommitInfo = buildCommit({
  commit: {
    id: '73fe17002aab4126a671c042678a62b2',
    repo: {name: 'images', project: {name: 'default'}},
  },
  description: 'initial commit',
  started: '2023-07-24T17:58:25Z',
  finished: '2023-07-25T21:45:05Z',
  sizeBytesUpperBound: '58644',
  details: {
    sizeBytes: '58644',
  },
});

export const IMAGE_COMMITS_NO_BRANCH: CommitInfo[] = [
  COMMIT_INFO_G2_NO_BRANCH,
  COMMIT_INFO_73_NO_BRANCH,
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
}: generateCommitsArgs): CommitInfo[] => {
  const commits: CommitInfo[] = [];

  for (let i = 0; i < n; i++) {
    commits.push(
      buildCommit({
        commit: {
          id: generateId(),
          repo: {name: repoName, project: {name: 'default'}},
          branch: {
            name: branchName,
            __typename: 'Branch',
          },
        },
        description: `item ${i}`,
        started: getISOStringFromUnix(i),
        finished: undefined,
        sizeBytesUpperBound: '0',
      }),
    );
  }

  for (let i = n - 2; i > 0; i--) {
    commits[i].parentCommit = commits[i + 1].commit;
  }

  return commits;
};

export const mockGetImageCommits = () =>
  rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (req, res, ctx) => {
      const body = await req.json();

      // responses from a specific global id filter or commit in path
      if (
        body.repo.project.name === 'default' &&
        body.repo.name === 'images' &&
        body.number === '2' &&
        body?.to?.id === COMMIT_INFO_C4.commit?.id
      ) {
        return res(ctx.json([COMMIT_INFO_C4]));
      }

      // responses from a specific global id filter or commit in path
      if (
        body.repo.project.name === 'default' &&
        body.repo.name === 'images' &&
        body.number === '2' &&
        body?.to?.id === COMMIT_INFO_4E.commit?.id
      ) {
        return res(ctx.json([COMMIT_INFO_4E]));
      }

      // response for the latest commit in a repo
      if (
        body.repo.project.name === 'default' &&
        body.repo.name === 'images' &&
        body.number === '2' &&
        !body?.to?.id
      ) {
        return res(ctx.json([COMMIT_INFO_4A]));
      }

      // regular listCommit response
      if (body.repo.name === 'images' && body.repo.project.name === 'default') {
        return res(ctx.json(IMAGE_COMMITS));
      }
      return res(ctx.json([]));
    },
  );

export const mockGetImageCommitsNoBranch = () =>
  rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (req, res, ctx) => {
      const body = await req.json();
      // regular listCommit response
      if (body.repo.name === 'images' && body.repo.project.name === 'default') {
        return res(ctx.json(IMAGE_COMMITS_NO_BRANCH));
      }
      return res(ctx.json([]));
    },
  );

export const mockEmptyCommits = () => {
  return rest.post<ListCommitRequest, Empty, CommitInfo[]>(
    '/api/pfs_v2.API/ListCommit',
    async (_req, res, ctx) => {
      return res(ctx.json([]));
    },
  );
};

export const mockEmptyCommitDiff = () =>
  rest.post<DiffFileRequest, Empty, DiffFileResponse[]>(
    '/api/pfs_v2.API/DiffFile',
    (_req, res, ctx) => {
      return res(ctx.json([]));
    },
  );

export const mockDiffFile = () => {
  return rest.post<DiffFileRequest, Empty, DiffFileResponse[]>(
    '/api/pfs_v2.API/DiffFile',
    async (_req, res, ctx) => {
      return res(
        ctx.json([
          {
            newFile: {
              file: {
                commit: {
                  repo: {
                    name: 'images',
                    type: 'user',
                    project: {
                      name: 'default',
                    },
                  },
                  id: '0b773ae5bd3248c2b84987df9c2385c8',
                  branch: {
                    repo: {
                      name: 'images',
                      type: 'user',
                      project: {
                        name: 'default',
                      },
                    },
                    name: 'master',
                  },
                },
                path: '/',
                datum: '',
              },
              fileType: FileType.DIR,
              committed: '2023-11-03T16:07:39.745857Z',
              sizeBytes: '58650',
            },
          },
        ]),
      );
    },
  );
};

export const inspectCommit = () => {
  return rest.post<InspectCommitRequest, Empty, CommitInfo | RequestError>(
    '/api/pfs_v2.API/InspectCommit',
    async (req, res, ctx) => {
      const body = await req.json();
      if (body.commit.id === '4a83c74809664f899261baccdb47cd90') {
        return res(ctx.json(COMMIT_INFO_4A));
      } else if (body.commit.id === 'c43fffd650a24b40b7d9f1bf90fcfdbe') {
        return res(ctx.json(COMMIT_INFO_C4));
      } else if (body.commit.id === '4eb1aa567dab483f93a109db4641ee75') {
        return res(ctx.json(COMMIT_INFO_4E));
      }

      return res(
        ctx.status(404),
        ctx.json({
          code: 5,
          message: 'not found',
          details: [],
        }),
      );
    },
  );
};

export const mockStartCommit = (id: string) =>
  rest.post<StartCommitRequest, Empty, Commit>(
    '/api/pfs_v2.API/StartCommit',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          repo: {
            name: 'images',
            type: 'user',
            project: {
              name: 'default',
            },
            __typename: 'Repo',
          },
          id,
          branch: {
            repo: {
              name: 'images',
              type: 'user',
              project: {
                name: 'default',
              },
            },
            name: 'master',
          },
          __typename: 'Commit',
        }),
      );
    },
  );

export const mockFinishCommit = () =>
  rest.post<FinishCommitRequest, Empty>(
    '/api/pfs_v2.API/FinishCommit',
    (_req, res, ctx) => {
      return res(ctx.json({}));
    },
  );
