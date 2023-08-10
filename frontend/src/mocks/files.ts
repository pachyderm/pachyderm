import {File, FileType, mockDeleteFilesMutation} from '@graphqlTypes';
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

export const mockDeleteFiles = () =>
  mockDeleteFilesMutation((_req, res, ctx) => {
    return res(ctx.data({deleteFiles: 'deleted'}));
  });
