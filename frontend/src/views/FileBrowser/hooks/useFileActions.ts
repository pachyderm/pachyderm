import {File} from '@graphqlTypes';
import {useState} from 'react';
import {useHistory} from 'react-router';

import useCurrentRepoWithLinkedPipeline from '@dash-frontend/hooks/useCurrentRepoWithLinkedPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getProxyEnabled} from '@dash-frontend/lib/runtimeVariables';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

import useArchiveDownload from './useArchiveDownload';
import useFileDisplay from './useFileDisplay';

const useFileActions = (
  file: File,
  openDeleteModal: () => void,
  includePreview = false,
) => {
  const [viewSource, setViewSource] = useState(false);
  const {copy, fileName, filePath, fileType} = useFileDisplay(file);
  const browserHistory = useHistory();
  const {loading: repoLoading, repo} = useCurrentRepoWithLinkedPipeline();
  const {repoId, branchId, projectId, commitId} = useUrlState();

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

  const {archiveDownload} = useArchiveDownload();

  const onMenuSelect = (id: string) => {
    switch (id) {
      case 'preview-file':
        return browserHistory.push(filePath);
      case 'copy-path':
        return copy();
      case 'download':
        return file.download
          ? window.open(file.download)
          : archiveDownload([file.path]);
      case 'delete':
        return openDeleteModal();
      default:
        return null;
    }
  };

  const toggleViewSource = () => setViewSource(!viewSource);

  const proxyEnabled = getProxyEnabled();

  let downloadText = 'Download';

  if (!file.download) {
    downloadText = proxyEnabled
      ? 'Download Zip'
      : 'Download (File too large to download)';
  }

  let iconItems: DropdownItem[] = [
    {
      id: 'copy-path',
      content: 'Copy Path',
      closeOnClick: true,
    },
    {
      id: 'download',
      content: downloadText,
      closeOnClick: true,
      disabled: !file.download && !proxyEnabled,
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
    viewSource,
    toggleViewSource,
    onMenuSelect,
    iconItems,
    fileType,
    branchId,
    handleBackNav,
  };
};

export default useFileActions;
