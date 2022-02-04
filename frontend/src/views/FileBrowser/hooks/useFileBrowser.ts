import {useModal} from '@pachyderm/components';
import {useState, useMemo, useCallback} from 'react';
import {useHistory} from 'react-router';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import {useFiles} from '@dash-frontend/hooks/useFiles';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

const useFileBrowser = () => {
  const browserHistory = useHistory();
  const {repoId, commitId, branchId, filePath, projectId} = useUrlState();
  const [fileFilter, setFileFilter] = useState('');
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

  const filteredFiles = useMemo(
    () =>
      files.filter(
        (file) =>
          !fileFilter ||
          file.path.toLowerCase().includes(fileFilter.toLowerCase()),
      ),
    [fileFilter, files],
  );

  const isDirectory = useMemo(() => {
    return path.endsWith('/');
  }, [path]);

  const fileToPreview = useMemo(() => {
    const hasFileType = path.slice(path.lastIndexOf('.') + 1) !== '';

    return hasFileType && files.find((file) => file.path === path);
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
  };
};

export default useFileBrowser;
