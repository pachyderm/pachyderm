import {File, FileType} from '@graphqlTypes';
import {useMemo} from 'react';

import {useFileDetails} from '@dash-frontend/components/CodePreview';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
import {useClipboardCopy} from '@pachyderm/components';

const useFileDisplay = (file: File) => {
  const {repoId, branchId, projectId} = useUrlState();
  const {parsedFilePath} = useFileDetails(file.path);
  const filePath = fileBrowserRoute({
    repoId,
    branchId,
    projectId,
    commitId: file.commitId,
    // remove forward slash from path for route
    filePath: file.path.slice(1),
  });

  const {copy} = useClipboardCopy(
    `${repoId}@${branchId}=${file.commitId}:${file.path}`,
  );

  const fileType = useMemo(() => {
    if (file.type === FileType.FILE) {
      return (parsedFilePath.ext || parsedFilePath.name)
        .toLowerCase()
        .replace('.', '');
    } else if (file.type === FileType.DIR) {
      return 'folder';
    } else {
      return 'unknown';
    }
  }, [file.type, parsedFilePath]);

  return {
    copy,
    fileName: parsedFilePath.base,
    filePath,
    fileType,
  };
};

export default useFileDisplay;
