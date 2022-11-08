import {useState, useMemo, useCallback} from 'react';
import {useHistory} from 'react-router';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import {useFiles} from '@dash-frontend/hooks/useFiles';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';
import {useModal} from '@pachyderm/components';

const useFileBrowser = () => {
  const browserHistory = useHistory();
  const {repoId, commitId, branchId, filePath, projectId} = useUrlState();
  const [fileFilter, setFileFilter] = useState('');
  const [diffOnly, setDiffOnly] = useLocalProjectSettings({
    projectId,
    key: 'diff_only_filebrowser',
  });
  const [fileView = 'list', setFileView] = useLocalProjectSettings({
    projectId,
    key: 'file_view',
  });
  const {closeModal, isOpen} = useModal(true);
  const {getPathFromFileBrowser} = useFileBrowserNavigation();
  const path = `/${filePath}`;

  const {files, loading} = useFiles({
    projectId,
    commitId,
    path: path || '/',
    branchName: branchId,
    repoName: repoId,
  });

  const filteredFiles = useMemo(() => {
    let filteredFiles = files?.files || [];

    if (diffOnly) {
      filteredFiles = filteredFiles.filter((file) => file.commitAction);
    }

    if (fileFilter) {
      filteredFiles = filteredFiles.filter((file) =>
        file.path.toLowerCase().includes(fileFilter.toLowerCase()),
      );
    }

    // sort updated/added files first
    if (fileView === 'icon') {
      filteredFiles = [...filteredFiles].sort((fileA, fileB) =>
        fileA.commitAction || '' > (fileB.commitAction || '') ? -1 : 1,
      );
    }

    return filteredFiles;
  }, [diffOnly, fileFilter, fileView, files?.files]);

  const isDirectory = useMemo(() => {
    return path.endsWith('/');
  }, [path]);

  const fileToPreview = useMemo(() => {
    const hasFileType = path.slice(path.lastIndexOf('.') + 1) !== '';

    return hasFileType && files?.files.find((file) => file.path === path);
  }, [path, files]);

  const handleHide = useCallback(() => {
    closeModal();

    setTimeout(
      () =>
        browserHistory.push(
          getPathFromFileBrowser(
            repoRoute({projectId, repoId, branchId}, false),
          ),
        ),
      500,
    );
  }, [
    projectId,
    repoId,
    branchId,
    browserHistory,
    closeModal,
    getPathFromFileBrowser,
  ]);

  return {
    fileFilter,
    setFileFilter,
    fileView,
    setFileView,
    closeModal,
    isOpen,
    filteredFiles,
    loading,
    fileToPreview,
    isDirectory,
    handleHide,
    files,
    setDiffOnly,
    diffOnly,
  };
};

export default useFileBrowser;
