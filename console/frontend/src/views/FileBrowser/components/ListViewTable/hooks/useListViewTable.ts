import {FileInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {useCallback, useEffect, useState} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import useArchiveDownload from '@dash-frontend/views/FileBrowser/hooks/useArchiveDownload';

const useListViewTable = (files: FileInfo[]) => {
  const {repoId, projectId, filePath} = useUrlState();
  const commitId = files[0]?.file?.commit?.id || '';

  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);

  const noFilesSelected = selectedFiles.length === 0;

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

  const {archiveDownload} = useArchiveDownload(projectId, repoId, commitId);

  const downloadSelected = useCallback(async () => {
    await archiveDownload(selectedFiles);
  }, [archiveDownload, selectedFiles]);

  return {
    repoId,
    selectedFiles,
    addSelection,
    downloadSelected,
    noFilesSelected,
  };
};

export default useListViewTable;
