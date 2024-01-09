import {useCallback, useEffect, useState} from 'react';
import {useHistory} from 'react-router';

import useDeleteFiles from '@dash-frontend/hooks/useDeleteFiles';
import {usePipeline} from '@dash-frontend/hooks/usePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import useArchiveDownload from '@dash-frontend/views/FileBrowser/hooks/useArchiveDownload';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useModal} from '@pachyderm/components';

const useListViewTable = () => {
  const {repoId, branchId, projectId, filePath} = useUrlState();
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const browserHistory = useHistory();

  const {loading: pipelineLoading, pipeline} = usePipeline({
    pipeline: {
      name: repoId,
      project: {name: projectId},
    },
  });

  const noFilesSelected = selectedFiles.length === 0;
  const isOutputRepo = Boolean(pipeline);

  const addSelection = useCallback(
    (filePath: string) => {
      if (selectedFiles.includes(filePath)) {
        setSelectedFiles((selectedFiles) =>
          selectedFiles.filter((file) => file !== filePath),
        );
      } else {
        setSelectedFiles((selectedFiles) => [...selectedFiles, filePath]);
      }
    },
    [selectedFiles],
  );

  useEffect(() => {
    setSelectedFiles([]);
  }, [filePath]);

  const {
    openModal: openDeleteModal,
    closeModal,
    isOpen: deleteModalOpen,
  } = useModal(false);

  const {
    deleteFiles: deleteHook,
    loading: deleteLoading,
    error: deleteError,
  } = useDeleteFiles({
    onSuccess: (id) => {
      browserHistory.push(
        fileBrowserRoute({
          repoId,
          branchId,
          projectId,
          commitId: id,
        }),
      );
    },
  });

  const deleteFiles = () => {
    deleteHook({
      filePaths: selectedFiles,
      repoId: repoId,
      branchId: branchId,
      projectId,
    });
  };

  const {archiveDownload} = useArchiveDownload();

  const downloadSelected = useCallback(async () => {
    await archiveDownload(selectedFiles);
  }, [archiveDownload, selectedFiles]);

  return {
    repoId,
    pipelineLoading,
    branchId,
    selectedFiles,
    addSelection,
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteFiles,
    deleteLoading,
    deleteError,
    downloadSelected,
    isOutputRepo,
    noFilesSelected,
  };
};

export default useListViewTable;
