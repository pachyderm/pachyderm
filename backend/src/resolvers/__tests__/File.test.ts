/* eslint-disable @typescript-eslint/naming-convention */
import {PUT_FILES_FROM_URLS_MUTATION} from '@dash-frontend/mutations/PutFilesFromURLs';
import {GET_FILES_QUERY} from '@dash-frontend/queries/GetFilesQuery';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {File, FileType, PutFilesFromUrLsMutation} from '@graphqlTypes';

const projectId = '1';

describe('File Resolver', () => {
  describe('files', () => {
    it('should return files for a given commit', async () => {
      const {data, errors = []} = await executeQuery<{files: File[]}>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
          },
        },
      );

      const files = data?.files;
      expect(errors?.length).toBe(0);
      expect(files?.length).toEqual(16);
      expect(files?.[0]?.path).toEqual('/AT-AT.png');
      expect(files?.[1]?.path).toEqual('/liberty.png');
      expect(files?.[2]?.path).toEqual('/cats/');

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

    it('should return files for a directory commit', async () => {
      const {data, errors = []} = await executeQuery<{files: File[]}>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/cats/',
            branchName: 'master',
            repoName: 'images',
          },
        },
      );

      const files = data?.files;
      expect(errors?.length).toBe(0);
      expect(files?.length).toEqual(1);
      expect(files?.[0]?.path).toEqual('/cats/kitten.png');
    });
  });
  describe('putFilesFromURLs', () => {
    it('should put files into the repo from the specified urls', async () => {
      const {data: files, errors = []} = await executeQuery<{files: File[]}>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
          },
        },
      );

      expect(errors?.length).toBe(0);
      expect(files?.files.length).toEqual(16);

      const {errors: mutationErrors = []} =
        await executeMutation<PutFilesFromUrLsMutation>(
          PUT_FILES_FROM_URLS_MUTATION,
          {
            args: {
              branch: 'master',
              repo: 'images',
              files: [
                {
                  url: 'https://imgur.com/a/rN7hjOQ',
                  path: '/puppy.png',
                },
              ],
              projectId,
            },
          },
        );
      expect(mutationErrors?.length).toBe(0);

      const {data: updatedFiles, errors: updatedErrors = []} =
        await executeQuery<{
          files: File[];
        }>(GET_FILES_QUERY, {
          args: {
            projectId,
            commitId: '1',
            path: '/',
            branchName: 'master',
            repoName: 'images',
          },
        });
      expect(updatedErrors.length).toBe(0);
      expect(updatedFiles?.files.length).toEqual(17);
    });
  });
});
