import merge from 'lodash/merge';
import {rest} from 'msw';

import {Empty, CODES} from '@dash-frontend/api/googleTypes';
import {
  EncodeArchiveUrlRequest,
  EncodeArchiveUrlResponse,
  FileInfo,
  FileType,
  ListFileRequest,
} from '@dash-frontend/api/pfs';
import {RequestError} from '@dash-frontend/api/utils/error';
export const buildFile = (file: FileInfo) => {
  const defaultFile: FileInfo = {
    __typename: 'FileInfo',
  };

  return merge(defaultFile, file);
};

export const MOCK_IMAGES_FILES: FileInfo[] = [
  buildFile({
    file: {
      commit: {
        repo: {
          name: 'images',
          type: 'user',
          project: {
            name: 'default',
          },
        },
        id: '4a83c74809664f899261baccdb47cd90',
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
      path: '/AT-AT.png',
      datum: 'default',
    },
    fileType: FileType.FILE,
    committed: '2023-11-08T18:12:19.363338Z',
    sizeBytes: '80590',
  }),
  buildFile({
    file: {
      commit: {
        repo: {
          name: 'images',
          type: 'user',
          project: {
            name: 'default',
          },
        },
        id: '4a83c74809664f899261baccdb47cd90',
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
      path: '/liberty.png',
      datum: 'default',
    },
    fileType: FileType.FILE,
    committed: '2023-11-08T18:12:19.363338Z',
    sizeBytes: '58650',
  }),
  buildFile({
    file: {
      commit: {
        repo: {
          name: 'images',
          type: 'user',
          project: {
            name: 'default',
          },
        },
        id: '4a83c74809664f899261baccdb47cd90',
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
      path: '/cats/',
      datum: 'default',
    },
    fileType: FileType.DIR,
    committed: '2023-11-08T18:12:19.363338Z',
    sizeBytes: '58650',
  }),
  buildFile({
    file: {
      commit: {
        repo: {
          name: 'images',
          type: 'user',
          project: {
            name: 'default',
          },
        },
        id: '4a83c74809664f899261baccdb47cd90',
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
      path: '/json_nested_arrays.json',
      datum: 'default',
    },
    fileType: FileType.FILE,
    committed: '2023-11-08T18:12:19.363338Z',
    sizeBytes: '200000001',
  }),
];

export const mockEmptyFiles = () =>
  rest.post<ListFileRequest, Empty, FileInfo[]>(
    '/api/pfs_v2.API/ListFile',
    (_req, res, ctx) => res(ctx.json([])),
  );

export const mockErrorFiles = () =>
  rest.post<ListFileRequest, never, RequestError>(
    '/api/pfs_v2.API/ListFile',
    (_req, res, ctx) =>
      res(
        ctx.status(500),
        ctx.json({
          code: CODES.Unknown,
          message: 'made up error',
        }),
      ),
  );

export const mockImagesFiles = () =>
  rest.post<ListFileRequest, Empty, FileInfo[]>(
    '/api/pfs_v2.API/ListFile',
    async (req, res, ctx) => {
      const body = await req.json();
      if (
        body.file.commit.branch.repo.project.name === 'default' &&
        body.file.commit.branch.repo.name === 'images' &&
        body.file.commit.branch.name === 'master' &&
        body.file.commit.id === '4a83c74809664f899261baccdb47cd90'
      ) {
        if (body.file.path === '/cats/') {
          return res(ctx.json([]));
        }
        return res(ctx.json(MOCK_IMAGES_FILES));
      }
      return res(ctx.json([]));
    },
  );

type generateFilesArgs = {
  n: number;
  repoName?: string;
  commitId?: string;
};

export const generatePagingFiles = ({
  n,
  repoName = 'repo',
  commitId = 'master',
}: generateFilesArgs) => {
  const commits: FileInfo[] = [];

  for (let id = 0; id < n; id++) {
    commits.push(
      buildFile({
        file: {
          commit: {
            repo: {
              name: repoName,
              type: 'user',
              project: {
                name: 'default',
              },
            },
            id: commitId,
            branch: {
              repo: {
                name: repoName,
                type: 'user',
                project: {
                  name: 'default',
                },
              },
              name: 'master',
            },
          },
          path: `/${id}.png`,
        },
        fileType: FileType.FILE,
        committed: '2023-11-08T18:12:19.363338Z',
        sizeBytes: '80590',
      }),
    );
  }
  return commits;
};

export const mockEncode = (files: string[]) =>
  rest.post<EncodeArchiveUrlRequest, Empty, EncodeArchiveUrlResponse>(
    '/encode/archive',
    async (req, res, ctx) => {
      const body = await req.json();
      if (
        body.projectId === 'default' &&
        body.repoId === 'images' &&
        body.commitId === '4a83c74809664f899261baccdb47cd90' &&
        JSON.stringify(body.paths) === JSON.stringify(files)
      ) {
        return res(
          ctx.json({
            url: '/proxyForward/archive/gCACFAwASxxgccGkVzh2qIdHkwutaaoJDgD.zip',
          }),
        );
      }
      return res(ctx.json({url: ''}));
    },
  );
