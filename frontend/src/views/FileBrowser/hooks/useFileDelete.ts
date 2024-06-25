import {useHistory} from 'react-router';

import {FileInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {useBranches} from '@dash-frontend/hooks/useBranches';
import useDeleteFiles from '@dash-frontend/hooks/useDeleteFiles';
import {usePipeline} from '@dash-frontend/hooks/usePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useModal} from '@pachyderm/components';

const useFileDelete = (file: FileInfo) => {
  const {
    openModal: openDeleteModal,
    closeModal,
    isOpen: deleteModalOpen,
  } = useModal(false);
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
  const commitBranch = branches?.find(
    (branch) => branch.head?.id === file.file?.commit?.id,
  );
  // TODO: FRON-1519
  const branchId = commitBranch?.branch?.name || undefined;

  const browserHistory = useHistory();

  const {
    deleteFiles: deleteHook,
    loading: deleteLoading,
    error,
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

  const deleteFile = () => {
    deleteHook({
      filePaths: [file.file?.path || ''],
      repoId: repoId,
      branchId: branchId || '',
      projectId,
    });
  };

  return {
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteFile,
    loading: deleteLoading,
    error,
    deleteDisabled:
      Boolean(pipeline) || pipelineLoading || branchesLoading || !commitBranch,
  };
};

export default useFileDelete;
