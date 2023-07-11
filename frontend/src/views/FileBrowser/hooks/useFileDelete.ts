import {File} from '@graphqlTypes';
import {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useDeleteFilesMutation} from '@dash-frontend/generated/hooks';
import useCurrentRepoWithLinkedPipeline from '@dash-frontend/hooks/useCurrentRepoWithLinkedPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useModal} from '@pachyderm/components';

const useFileDelete = (file: File) => {
  const {
    openModal: openDeleteModal,
    closeModal,
    isOpen: deleteModalOpen,
  } = useModal(false);
  const {repoId, branchId, projectId} = useUrlState();
  const [deleteFilesMutation, {loading: deleteLoading, error}] =
    useDeleteFilesMutation();
  const {loading: repoLoading, repo} = useCurrentRepoWithLinkedPipeline();
  const browserHistory = useHistory();

  const deleteFile = useCallback(async () => {
    const deleteCommit = await deleteFilesMutation({
      variables: {
        args: {
          filePaths: [file.path],
          repo: repoId,
          branch: branchId,
          projectId,
        },
      },
    });
    deleteCommit.data?.deleteFiles &&
      browserHistory.push(
        fileBrowserRoute({
          repoId,
          branchId,
          projectId,
          commitId: deleteCommit.data?.deleteFiles,
        }),
      );
  }, [
    branchId,
    browserHistory,
    deleteFilesMutation,
    file.path,
    projectId,
    repoId,
  ]);

  return {
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteFile,
    loading: deleteLoading,
    error,
    deleteDisabled: Boolean(repo?.linkedPipeline) || repoLoading,
  };
};

export default useFileDelete;
