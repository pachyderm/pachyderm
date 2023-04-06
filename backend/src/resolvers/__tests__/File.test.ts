import {PUT_FILES_FROM_URLS_MUTATION} from '@dash-frontend/mutations/PutFilesFromURLs';
import {GET_FILES_QUERY} from '@dash-frontend/queries/GetFilesQuery';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  FileQueryResponse,
  FileType,
  GetFilesQuery,
  PutFilesFromUrLsMutation,
} from '@graphqlTypes';

const projectId = 'Solar-Panel-Data-Sorting';

describe('File Resolver', () => {
  describe('files', () => {
    it('should return files for a given commit', async () => {
      const {data, errors = []} = await executeQuery<{
        files: FileQueryResponse;
      }>(GET_FILES_QUERY, {
        args: {
          projectId,
          commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
          path: '/',
          branchName: 'master',
          repoName: 'images',
        },
      });

      const files = data?.files.files;
      expect(errors).toHaveLength(0);
      expect(files).toHaveLength(17);
      expect(data?.files.diff?.size).toBe(58644);
      expect(data?.files.diff?.sizeDisplay).toBe('58.65 kB');
      expect(data?.files.diff?.filesAdded.count).toBe(1);
      expect(data?.files.diff?.filesUpdated.count).toBe(0);
      expect(data?.files.diff?.filesDeleted.count).toBe(0);
      expect(data?.files.diff?.filesAdded.sizeDelta).toBe(58644);
      expect(files?.[0]?.path).toBe('/AT-AT.png');
      expect(files?.[1]?.path).toBe('/liberty.png');
      expect(files?.[2]?.path).toBe('/cats/');

      expect(files?.[0]?.commitId).toBe('d350c8d08a644ed5b2ee98c035ab6b33');
      expect(files?.[0]?.committed).toStrictEqual({
        __typename: 'Timestamp',
        nanos: 0,
        seconds: 1614126189,
      });
      expect(files?.[0]?.download).toContain(
        '/download/undefined/images/master/d350c8d08a644ed5b2ee98c035ab6b33/AT-AT.png', // TODO; mock server not implementing project route?
      );
      expect(files?.[0]?.downloadDisabled).toBe(false);
      expect(files?.[0]?.hash).toBe(
        'P2fxZjakvux5dNsEfc0iCx1n3Kzo2QcDlzu9y3Ra1gc=',
      );
      expect(files?.[0]?.repoName).toBe('images');
      expect(files?.[0]?.sizeBytes).toBe(80588);
      expect(files?.[0]?.sizeDisplay).toBe('80.59 kB');
      expect(files?.[0]?.type).toBe(FileType.FILE);
      expect(files?.[1]?.commitAction).toBe('ADDED');
    });

    it('should return diff data for updated files in a commit', async () => {
      const {data, errors = []} = await executeQuery<{
        files: FileQueryResponse;
      }>(GET_FILES_QUERY, {
        args: {
          projectId: 'Data-Cleaning-Process',
          path: '/',
          branchName: 'master',
          repoName: 'images',
        },
      });

      const files = data?.files.files;
      expect(errors?.length).toBe(0);
      expect(files?.length).toBe(3);
      expect(data?.files.diff?.size).toBe(10000);
      expect(data?.files.diff?.sizeDisplay).toBe('10 kB');
      expect(data?.files.diff?.filesAdded.count).toBe(0);
      expect(data?.files.diff?.filesDeleted.count).toBe(0);
      expect(data?.files.diff?.filesUpdated.count).toBe(1);
      expect(data?.files.diff?.filesUpdated.sizeDelta).toBe(10000);
    });

    it('should return diff data for deleted files in a commit', async () => {
      const {data, errors = []} = await executeQuery<{
        files: FileQueryResponse;
      }>(GET_FILES_QUERY, {
        args: {
          projectId: 'Solar-Price-Prediction-Modal',
          path: '/',
          branchName: 'master',
          repoName: 'images',
        },
      });

      const files = data?.files.files;
      expect(errors?.length).toBe(0);
      expect(files?.length).toBe(3);
      expect(data?.files.diff?.size).toBe(-50588);
      expect(data?.files.diff?.sizeDisplay).toBe('-50.59 kB');
      expect(data?.files.diff?.filesAdded.count).toBe(0);
      expect(data?.files.diff?.filesUpdated.count).toBe(0);
      expect(data?.files.diff?.filesDeleted.count).toBe(1);
      expect(data?.files.diff?.filesDeleted.sizeDelta).toBe(-50588);
    });

    it('should return files for a directory commit', async () => {
      const {data, errors = []} = await executeQuery<{
        files: FileQueryResponse;
      }>(GET_FILES_QUERY, {
        args: {
          projectId,
          commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
          path: '/cats/',
          branchName: 'master',
          repoName: 'images',
        },
      });

      const files = data?.files.files;
      expect(errors).toHaveLength(0);
      expect(files).toHaveLength(2);
      expect(files?.[0]?.path).toBe('/cats/kitten.png');
      expect(files?.[1]?.path).toBe('/cats/test.png');
      expect(files?.[1]?.downloadDisabled).toBe(true);
    });
  });

  describe('files paging', () => {
    it('should return the first page if no cursor is specified', async () => {
      const {data, errors = []} = await executeQuery<GetFilesQuery>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
            limit: 3,
            reverse: false,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(3);
      expect(data?.files.cursor).toBe('/cats/');
      expect(data?.files.hasNextPage).toBe(true);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
          path: '/AT-AT.png',
          repoName: 'images',
        }),
      );
    });

    it('should return the next page if cursor is specified', async () => {
      const {data, errors = []} = await executeQuery<GetFilesQuery>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
            cursorPath: '/cats/',
            limit: 3,
            reverse: false,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(3);
      expect(data?.files.cursor).toBe('/csv_tabs.csv');
      expect(data?.files.hasNextPage).toBe(true);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '531f844bd184e913b050d49856e8d438',
          path: '/carriers_list.textpb',
          repoName: 'samples',
        }),
      );
    });

    it('should return no cursor if there are no more jobs after the requested page', async () => {
      const {data, errors = []} = await executeQuery<GetFilesQuery>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
            cursorPath: '/json_object_array.json',
            limit: 10,
            reverse: false,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(7);
      expect(data?.files.cursor).toBeNull();
      expect(data?.files.hasNextPage).toBe(false);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '531f844bd184e913b050d49856e8d438',
          path: '/json_single_field.json',
          repoName: 'samples',
        }),
      );
    });

    it('should return the reverse order is reverse is true', async () => {
      const {data, errors = []} = await executeQuery<GetFilesQuery>(
        GET_FILES_QUERY,
        {
          args: {
            projectId,
            commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
            path: '/',
            branchName: 'master',
            repoName: 'images',
            limit: 1,
            reverse: true,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(1);
      expect(data?.files.cursor).toBe('/yml_spec.yml');
      expect(data?.files.hasNextPage).toBe(true);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '531f844bd184e913b050d49856e8d438',
          path: '/yml_spec.yml',
          repoName: 'samples',
        }),
      );
    });
  });
  describe('putFilesFromURLs', () => {
    it('should put files into the repo from the specified urls', async () => {
      const {data, errors = []} = await executeQuery<{
        files: FileQueryResponse;
      }>(GET_FILES_QUERY, {
        args: {
          projectId,
          commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
          path: '/',
          branchName: 'master',
          repoName: 'images',
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(17);

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
      expect(mutationErrors).toHaveLength(0);

      const {data: updatedFiles, errors: updatedErrors = []} =
        await executeQuery<{
          files: FileQueryResponse;
        }>(GET_FILES_QUERY, {
          args: {
            projectId,
            commitId: '1',
            path: '/',
            branchName: 'master',
            repoName: 'images',
          },
        });
      expect(updatedErrors).toHaveLength(0);
      expect(updatedFiles?.files.files).toHaveLength(18);
    });
  });
});
