import {DELETE_FILES_MUTATION} from '@dash-frontend/mutations/DeleteFiles';
import {PUT_FILES_FROM_URLS_MUTATION} from '@dash-frontend/mutations/PutFilesFromURLs';
import {GET_FILE_DOWNLOAD_QUERY} from '@dash-frontend/queries/GetFileDownloadQuery';
import {GET_FILES_QUERY} from '@dash-frontend/queries/GetFilesQuery';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  DeleteFilesMutation,
  FileQueryResponse,
  FileType,
  GetFilesQuery,
  FileDownloadQuery,
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
      expect(files).toHaveLength(20);
      expect(files?.[0]?.path).toBe('/AT-AT.png');
      expect(files?.[1]?.path).toBe('/liberty.png');
      expect(files?.[2]?.path).toBe('/cats/');

      expect(files?.[0]?.commitId).toBe('0918ac9d5daa76b86e3bb5e88e4c43a4');
      expect(files?.[0]?.committed).toStrictEqual({
        __typename: 'Timestamp',
        nanos: 0,
        seconds: 1614126189,
      });
      expect(files?.[0]?.download).toContain(
        '/download/undefined/images/master/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png', // TODO; mock server not implementing project route?
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
      expect(data?.files.cursor).toBe('/carriers_list.textpb');
      expect(data?.files.hasNextPage).toBe(true);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '0918ac9d5daa76b86e3bb5e88e4c43a4',
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
            cursorPath: '/carriers_list.textpb',
            limit: 3,
            reverse: false,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(3);
      expect(data?.files.cursor).toBe('/html_pachyderm.html');
      expect(data?.files.hasNextPage).toBe(true);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '9d5daa0918ac4c43a476b86e3bb5e88e',
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
            cursorPath: '/jsonl_people.jsonl',
            limit: 10,
            reverse: false,
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.files.files).toHaveLength(8);
      expect(data?.files.cursor).toBeNull();
      expect(data?.files.hasNextPage).toBe(false);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '9d5daa0918ac4c43a476b86e3bb5e88e',
          path: '/jsonl_people.jsonl',
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
      expect(data?.files.cursor).toBe('/file.unknown');
      expect(data?.files.hasNextPage).toBe(true);
      expect(data?.files.files[0]).toEqual(
        expect.objectContaining({
          __typename: 'File',
          commitId: '9d5daa0918ac4c43a476b86e3bb5e88e',
          path: '/yml_spec_too_large.yml',
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
      expect(data?.files.files).toHaveLength(20);

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
      expect(updatedFiles?.files.files).toHaveLength(21);
    });
  });

  describe('file download', () => {
    it('should return a valid url for root directory', async () => {
      const {data, errors = []} = await executeQuery<FileDownloadQuery>(
        GET_FILE_DOWNLOAD_QUERY,
        {
          args: {
            projectId: 'default',
            repoId: 'edges',
            commitId: '3f549092055c471980559f9062098ab4',
            paths: ['/'],
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.fileDownload).toBe(
        '/archive/ASi1L_0gMIEBAGRlZmF1bHQvZWRnZXNAM2Y1NDkwOTIwNTVjNDcxOTgwNTU5ZjkwNjIwOThhYjQ6Lw.zip',
      );
    });
    it('should return a valid url for multiple files in a directory', async () => {
      const {data, errors = []} = await executeQuery<FileDownloadQuery>(
        GET_FILE_DOWNLOAD_QUERY,
        {
          args: {
            projectId: 'default',
            repoId: 'images',
            commitId: 'master',
            paths: ['/folder/folder/pic.png', '/kitten.png', '/AT-AT.png'],
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.fileDownload).toBe(
        '/archive/ASi1L_0gbw0CAEQDZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi9BVC1BVC5wbmcAZm9sZGVycGlja2l0dGVuLnBuZwMAULGhQpSQWj0B.zip',
      );
    });
  });

  describe('deleteFiles', () => {
    it('should delete requested files', async () => {
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
      expect(data?.files.files).toHaveLength(20);

      const {errors: mutationErrors = []} =
        await executeMutation<DeleteFilesMutation>(DELETE_FILES_MUTATION, {
          args: {
            branch: 'master',
            repo: 'images',
            filePaths: ['/liberty.png', '/csv_commas.csv'],
            projectId,
          },
        });
      expect(mutationErrors).toHaveLength(0);

      const {data: updatedFiles, errors: updatedErrors = []} =
        await executeQuery<{
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
      expect(updatedErrors).toHaveLength(0);
      expect(updatedFiles?.files.files).toHaveLength(18);
    });
  });
});
