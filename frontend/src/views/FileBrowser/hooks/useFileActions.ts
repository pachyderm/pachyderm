import {File} from '@graphqlTypes';
import {useState} from 'react';
import {useHistory} from 'react-router';

import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

import useFileDisplay from './useFileDisplay';

const useFileActions = (
  file: File,
  openDeleteModal: () => void,
  includePreview = false,
) => {
  const [viewSource, setViewSource] = useState(false);
  const {
    copy,
    fileName,
    filePath,
    previewSupported,
    fileMajorType,
    fileType,
    viewSourceSupported,
  } = useFileDisplay(file);
  const browserHistory = useHistory();
  const {loading: repoLoading, repo} = useCurrentRepo();
  const {repoId, branchId, projectId, commitId} = useUrlState();

  const download =
    !file.downloadDisabled && file.download ? file.download : undefined;

  const deleteDisabled = Boolean(repo?.linkedPipeline) || repoLoading;

  const handleBackNav = () => {
    const filePaths = file.path.split('/');
    filePaths.pop();
    const parentPath = filePaths.join('/').concat('/');
    browserHistory.push(
      fileBrowserRoute({
        repoId,
        branchId,
        projectId,
        commitId,
        filePath: parentPath === '/' ? undefined : parentPath,
      }),
    );
  };

  const onMenuSelect = (id: string) => {
    switch (id) {
      case 'preview-file':
        return browserHistory.push(filePath);
      case 'copy-path':
        return copy();
      case 'download':
        return download ? window.open(download) : null;
      case 'delete':
        return openDeleteModal();
      default:
        return null;
    }
  };

  const toggleViewSource = () => setViewSource(!viewSource);

  let iconItems: DropdownItem[] = [
    {
      id: 'copy-path',
      content: 'Copy Path',
      closeOnClick: true,
    },
    {
      id: 'download',
      content: file.downloadDisabled
        ? 'Download (File too large to download)'
        : 'Download',
      closeOnClick: true,
      disabled: !!file.downloadDisabled,
    },
  ];

  if (includePreview) {
    iconItems = [
      {
        id: 'preview-file',
        content: 'Preview',
        closeOnClick: true,
      },
      ...iconItems,
    ];
  }

  if (!deleteDisabled) {
    iconItems.push({
      id: 'delete',
      content: 'Delete',
      closeOnClick: true,
    });
  }

  return {
    fileName,
    filePath,
    previewSupported,
    viewSourceSupported,
    viewSource,
    toggleViewSource,
    onMenuSelect,
    iconItems,
    fileMajorType,
    fileType,
    branchId,
    handleBackNav,
  };
};

export default useFileActions;
