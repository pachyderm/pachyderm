import {
  File,
  FileCommitState,
  GetFilesQuery,
  mockGetFilesQuery,
  FileType,
  mockDeleteFilesMutation,
} from '@graphqlTypes';
import merge from 'lodash/merge';

export const buildFile = (file: Partial<File>): File => {
  const defaultFile = {
    committed: null,
    commitId: 'default',
    download: null,
    hash: '',
    path: '',
    repoName: 'images',
    sizeBytes: 0,
    sizeDisplay: '0 B',
    type: FileType.FILE,
    commitAction: null,
    __typename: 'File',
  };

  return merge(defaultFile, file);
};

export const MOCK_EMPTY_FILES: GetFilesQuery = {
  files: {
    files: [],
    cursor: null,
    hasNextPage: false,
    __typename: 'PageableFile',
  },
};

export const MOCK_IMAGES_FILES: File[] = [
  buildFile({
    commitId: '4a83c74809664f899261baccdb47cd90',
    repoName: 'images',
    path: '/AT-AT.png',
    sizeBytes: 80590,
    sizeDisplay: '80.59 kB',
  }),
  buildFile({
    commitId: '4a83c74809664f899261baccdb47cd90',
    repoName: 'images',
    path: '/liberty.png',
    commitAction: FileCommitState.ADDED,
    sizeBytes: 58650,
    sizeDisplay: '58.65 kB',
  }),
  buildFile({
    commitId: '4a83c74809664f899261baccdb47cd90',
    repoName: 'images',
    path: '/cats/',
  }),
];

export const mockEmptyFiles = () =>
  mockGetFilesQuery((_req, res, ctx) => {
    return res(ctx.data(MOCK_EMPTY_FILES));
  });

export const mockErrorFiles = () =>
  mockGetFilesQuery((_req, res, ctx) => {
    return res(ctx.errors(['error retrieving files']));
  });

export const mockImagesFiles = () =>
  mockGetFilesQuery((req, res, ctx) => {
    const {projectId, repoName, branchName, commitId, path} =
      req.variables.args;

    if (
      projectId === 'default' &&
      repoName === 'images' &&
      branchName === 'master' &&
      commitId === '4a83c74809664f899261baccdb47cd90'
    ) {
      if (path === '/cats/') {
        return res(
          ctx.data({
            files: {
              files: [],
              cursor: null,
              hasNextPage: false,
              __typename: 'PageableFile',
            },
          }),
        );
      }
      return res(
        ctx.data({
          files: {
            files: MOCK_IMAGES_FILES,
            cursor: null,
            hasNextPage: false,
            __typename: 'PageableFile',
          },
        }),
      );
    }
    return res();
  });

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
  const commits: File[] = [];

  for (let id = 0; id < n; id++) {
    commits.push(
      buildFile({commitId: commitId, repoName: repoName, path: `/${id}.png`}),
    );
  }
  return commits;
};

export const mockDeleteFiles = () =>
  mockDeleteFilesMutation((_req, res, ctx) => {
    return res(ctx.data({deleteFiles: 'deleted'}));
  });
