import {useModal} from '@pachyderm/components';
import {useState, useMemo, useCallback} from 'react';

import {useFiles} from '@dash-frontend/hooks/useFiles';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

const useFileBrowser = () => {
  const {repoId, commitId, branchId, filePath, projectId} = useUrlState();
  const [fileFilter, setFileFilter] = useState('');
  const [fileView = 'list', setFileView] = useLocalProjectSettings({
    projectId,
    key: 'file_view',
  });
  const {closeModal, isOpen} = useModal(true);
  const {viewState, setUrlFromViewState} = useUrlQueryState();
  const path = `/${filePath}`;

  const {files, loading} = useFiles({
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

    const nextPath =
      viewState.prevFileBrowserPath ||
      repoRoute({projectId, repoId, branchId}, false);

    setTimeout(
      () => setUrlFromViewState({prevFileBrowserPath: undefined}, nextPath),
      500,
    );
  }, [projectId, repoId, branchId, setUrlFromViewState, closeModal, viewState]);

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
