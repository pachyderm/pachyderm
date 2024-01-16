import {useState} from 'react';
import {useHistory} from 'react-router';

import {FileInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {usePipeline} from '@dash-frontend/hooks/usePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getDownloadLink} from '@dash-frontend/lib/fileUtils';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

import useArchiveDownload from './useArchiveDownload';
import useFileDisplay from './useFileDisplay';

const useFileActions = (
  file: FileInfo,
  openDeleteModal: () => void,
  includePreview = false,
) => {
  const [viewSource, setViewSource] = useState(false);
  const {copy, fileName, filePath, fileType} = useFileDisplay(file);
  const browserHistory = useHistory();

  const {repoId, projectId} = useUrlState();
  const commitId = file.file?.commit?.id || '';
  const branchId = file.file?.commit?.branch?.name;

  const {loading: pipelineLoading, pipeline} = usePipeline({
    pipeline: {
      name: repoId,
      project: {name: projectId},
    },
  });
  const deleteDisabled = Boolean(pipeline) || pipelineLoading || !branchId;

  const downloadLink = getDownloadLink(file);

  const handleBackNav = () => {
    const filePaths = file.file?.path ? file.file.path.split('/') : [];
    filePaths.pop();
    const parentPath = filePaths.join('/').concat('/');
    browserHistory.push(
      fileBrowserRoute({
        repoId,
        projectId,
        commitId,
        filePath: parentPath === '/' ? undefined : parentPath,
      }),
    );
  };

  const {archiveDownload} = useArchiveDownload(projectId, repoId, commitId);

  const onMenuSelect = (id: string) => {
    switch (id) {
      case 'preview-file':
        return browserHistory.push(filePath);
      case 'copy-path':
        return copy();
      case 'download':
        return downloadLink
          ? window.open(downloadLink)
          : archiveDownload([file.file?.path || '']);
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
      content: downloadLink ? 'Download' : 'Download Zip',
      closeOnClick: true,
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
