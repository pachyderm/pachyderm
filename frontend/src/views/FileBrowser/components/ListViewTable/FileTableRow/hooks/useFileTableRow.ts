import {File} from '@graphqlTypes';
import {useHistory} from 'react-router';

import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import {DropdownItem} from '@pachyderm/components';

import useFileDisplay from '../../../../hooks/useFileDisplay';

const useFileTableRow = (file: File, openDeleteModal: () => void) => {
  const {copy, fileName, filePath, previewSupported} = useFileDisplay(file);
  const browserHistory = useHistory();
  const {loading: repoLoading, repo} = useCurrentRepo();

  const download =
    !file.downloadDisabled && file.download ? file.download : undefined;

  const deleteDisabled = Boolean(repo?.linkedPipeline) || repoLoading;

  const onOverflowMenuSelect = (id: string) => {
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

  const getPreviewActionText = () => {
    if (!previewSupported) {
      return 'Preview (File type not supported)';
    }
    if (file.downloadDisabled) {
      return 'Preview (File too large to preview)';
    }
    return 'Preview';
  };

  const iconItems: DropdownItem[] = [
    {
      id: 'preview-file',
      content: getPreviewActionText(),
      closeOnClick: true,
      disabled: !previewSupported || !!file.downloadDisabled,
    },
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
    onOverflowMenuSelect,
    iconItems,
  };
};

export default useFileTableRow;
