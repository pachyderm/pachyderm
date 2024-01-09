import {useMemo} from 'react';

import {useFileDetails} from '@dash-frontend/components/CodePreview';
import {FileInfo, FileType} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useClipboardCopy} from '@pachyderm/components';

const useFileDisplay = (file: FileInfo) => {
  const {repoId, branchId, projectId} = useUrlState();
  const {parsedFilePath} = useFileDetails(file.file?.path || '');
  const filePath = fileBrowserRoute({
    repoId,
    branchId,
    projectId,
    commitId: file.file?.commit?.id || '',
    // remove forward slash from path for route
    filePath: file.file?.path ? file.file?.path.slice(1) : '',
  });

  const {copy} = useClipboardCopy(
    `${repoId}@${branchId}=${file.file?.commit?.id || ''}:${file.file?.path}`,
  );

  const fileType = useMemo(() => {
    if (file.fileType === FileType.FILE) {
      return (parsedFilePath.ext || parsedFilePath.name)
        .toLowerCase()
        .replace('.', '');
    } else if (file.fileType === FileType.DIR) {
      return 'folder';
    } else {
      return 'unknown';
    }
  }, [file.fileType, parsedFilePath]);

  return {
    copy,
    fileName: parsedFilePath.base,
    filePath,
    fileType,
  };
};

export default useFileDisplay;
