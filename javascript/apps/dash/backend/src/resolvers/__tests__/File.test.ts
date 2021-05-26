/* eslint-disable @typescript-eslint/naming-convention */
import {executeQuery} from '@dash-backend/testHelpers';
import {GET_FILES_QUERY} from '@dash-frontend/queries/GetFilesQuery';
import {File, FileType} from '@graphqlTypes';

describe('File Resolver', () => {
  describe('files', () => {
    it('should return files for a given commit', async () => {
      const {data, errors = []} = await executeQuery<{files: File[]}>(
        GET_FILES_QUERY,
        {
          args: {
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
          },
        },
      );

      const files = data?.files;
      expect(errors?.length).toBe(0);
      expect(files?.length).toEqual(4);
      expect(files?.[0]?.path).toEqual('/AT-AT.png');
      expect(files?.[1]?.path).toEqual('/liberty.png');
      expect(files?.[2]?.path).toEqual('/kitten.png');
      expect(files?.[3]?.path).toEqual('/cats/');

      expect(files?.[0]?.commitId).toBe('d350c8d08a644ed5b2ee98c035ab6b33');
      expect(files?.[0]?.committed).toStrictEqual({
        __typename: 'Timestamp',
        nanos: 0,
        seconds: 1614126189,
      });
      expect(files?.[0]?.download).toContain(
        '/download/images/master/d350c8d08a644ed5b2ee98c035ab6b33/AT-AT.png',
      );
      expect(files?.[0]?.downloadDisabled).toBe(false);
      expect(files?.[0]?.hash).toBe(
        'P2fxZjakvux5dNsEfc0iCx1n3Kzo2QcDlzu9y3Ra1gc=',
      );
      expect(files?.[0]?.repoName).toBe('images');
      expect(files?.[0]?.sizeBytes).toBe(80588);
      expect(files?.[0]?.sizeDisplay).toBe('78.7 KB');
      expect(files?.[0]?.type).toBe(FileType.FILE);
    });
  });
});
