import {useMutation} from '@tanstack/react-query';

import {finishCommit, modifyFile, startCommit} from '@dash-frontend/api/pfs';
import {ModifyFileRequest} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
export type DeleteFilesRequest = {
  projectId: string;
  repoId: string;
  branchId: string;
  filePaths: string[];
};

type useDeleteFilesArgs = {
  onSuccess?: (id: string) => void;
  onError?: (error: Error) => void;
};

const bodyBuilder = (modifyFileRequests: ModifyFileRequest[]): string => {
  return modifyFileRequests
    .map((modifyFileRequest) => JSON.stringify(modifyFileRequest))
    .join('\n');
};

const useDeleteFiles = ({onSuccess, onError}: useDeleteFilesArgs) => {
  const {
    mutate: deleteFilesMutation,
    isPending: loading,
    error,
    data,
  } = useMutation({
    mutationKey: ['deleteFiles'],
    mutationFn: async (req: DeleteFilesRequest) => {
      const {projectId, repoId, branchId, filePaths} = req;

      const deleteCommit = await startCommit({
        branch: {
          name: branchId,
          repo: {
            name: repoId,
            type: 'user',
            project: {
              name: projectId,
            },
          },
        },
      });

      // Modify file requests need to start with a setCommit object
      // referencing an open commit.
      try {
        await modifyFile(
          bodyBuilder([
            {
              setCommit: deleteCommit,
            },
            ...filePaths.map((path) => ({
              deleteFile: {
                path,
              },
            })),
          ]),
        );
      } finally {
        await finishCommit({
          commit: deleteCommit,
        });
      }
      return deleteCommit.id || '';
    },
    onSuccess: (data) => onSuccess && onSuccess(data),
    onError: (error) => onError && onError(error),
  });

  return {
    deleteFilesResponse: data,
    deleteFiles: deleteFilesMutation,
    error: getErrorMessage(error),
    loading,
  };
};

export default useDeleteFiles;
