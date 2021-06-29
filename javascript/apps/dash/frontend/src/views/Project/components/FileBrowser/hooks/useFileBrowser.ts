import {useModal} from '@pachyderm/components';
import {useState, useMemo} from 'react';

import {useFiles} from '@dash-frontend/hooks/useFiles';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useFileBrowser = () => {
  const {repoId, commitId, branchId, filePath} = useUrlState();
  const [fileFilter, setFileFilter] = useState('');
  const [fileView, setFileView] = useState<'list' | 'icon'>('list');
  const {closeModal, isOpen} = useModal(true);
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

  const fileToPreview = useMemo(() => {
    const hasFileType = path.slice(path.lastIndexOf('.') + 1);
    return hasFileType && files.find((file) => file.path === path);
  }, [path, files]);

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
  };
};

export default useFileBrowser;
