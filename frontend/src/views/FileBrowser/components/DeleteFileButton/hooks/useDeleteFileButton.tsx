import {File} from '@graphqlTypes';
import {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useDeleteFileMutation} from '@dash-frontend/generated/hooks';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useModal} from '@pachyderm/components';

const useDeleteFileButton = (file: File) => {
  const {openModal, closeModal, isOpen} = useModal(false);
  const {repoId, branchId, projectId} = useUrlState();
  const [deleteFileMutation, {loading: deleteLoading, error}] =
    useDeleteFileMutation();
  const {loading: repoLoading, repo} = useCurrentRepo();
  const browserHistory = useHistory();

  const deleteFile = useCallback(async () => {
    const deleteCommit = await deleteFileMutation({
      variables: {
        args: {
          filePath: file.path,
          repo: repoId,
          branch: branchId,
          projectId,
        },
      },
    });
    deleteCommit.data?.deleteFile &&
      browserHistory.push(
        fileBrowserRoute({
          repoId,
          branchId,
          projectId,
          commitId: deleteCommit.data?.deleteFile,
        }),
      );
  }, [
    branchId,
    browserHistory,
    deleteFileMutation,
    file.path,
    projectId,
    repoId,
  ]);

  return {
    isOpen,
    openModal,
    closeModal,
    deleteFile,
    loading: deleteLoading,
    error,
    deleteDisabled: Boolean(repo?.linkedPipeline) || repoLoading,
  };
};

export default useDeleteFileButton;
