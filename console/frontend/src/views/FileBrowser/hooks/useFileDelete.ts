import {FileInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useBranches} from '@dash-frontend/hooks/useBranches';
import useDeleteFiles from '@dash-frontend/hooks/useDeleteFiles';
import {usePipeline} from '@dash-frontend/hooks/usePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useModal} from '@pachyderm/components';

type BranchFormFields = {
  branch: string;
};

const useFileDelete = (files: FileInfo[], selectedFiles?: string[]) => {
  const {
    openModal: openDeleteConfirmationModal,
    closeModal: closeDeleteConfirmationModal,
    isOpen: deleteConfirmationModalOpen,
  } = useModal(false);
  const {
    openModal: openBranchSelectionModal,
    closeModal: closeBranchSelectionModal,
    isOpen: branchSelectionModalOpen,
  } = useModal(false);

  const browserHistory = useHistory();
  const {repoId, projectId} = useUrlState();
  const {loading: pipelineLoading, pipeline} = usePipeline({
    pipeline: {
      name: repoId,
      project: {name: projectId},
    },
  });
  const {branches, loading: branchesLoading} = useBranches({
    projectId,
    repoId,
  });
  const commitBranches = branches?.filter(
    (branch) => branch.head?.id === files[0].file?.commit?.id,
  );

  const firstBranchId =
    (commitBranches &&
      commitBranches?.length > 0 &&
      commitBranches[0]?.branch?.name) ||
    undefined;
  const hasManyBranches = commitBranches && commitBranches?.length > 1;

  const {
    deleteFiles: deleteHook,
    loading: deleteLoading,
    error: deleteError,
  } = useDeleteFiles({
    onSuccess: (id) => {
      browserHistory.push(
        fileBrowserRoute({
          repoId,
          projectId,
          commitId: id,
        }),
      );
    },
  });

  const deleteFiles = useCallback(
    (selectedBranch?: string) => {
      deleteHook({
        filePaths: selectedFiles || files.map((file) => file.file?.path || ''),
        repoId: repoId,
        branchId: selectedBranch || firstBranchId || '',
        projectId,
      });
    },
    [deleteHook, files, firstBranchId, projectId, repoId, selectedFiles],
  );

  const submitBranchSelectionForm = useCallback(
    (formData: BranchFormFields) => {
      deleteFiles(formData.branch);
    },
    [deleteFiles],
  );

  const isOutputRepo = Boolean(pipeline);
  const deleteDisabled =
    isOutputRepo ||
    pipelineLoading ||
    branchesLoading ||
    !commitBranches ||
    commitBranches.length === 0;

  return {
    openDeleteConfirmationModal,
    closeDeleteConfirmationModal,
    deleteConfirmationModalOpen,
    openBranchSelectionModal,
    closeBranchSelectionModal,
    branchSelectionModalOpen,
    submitBranchSelectionForm,
    deleteFiles: () => deleteFiles(),
    deleteLoading,
    deleteError,
    commitBranches,
    hasManyBranches,
    deleteDisabled,
    isOutputRepo,
    firstBranchId,
  };
};

export default useFileDelete;
