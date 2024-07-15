import {useCallback} from 'react';

import {encodeArchiveUrl} from '@dash-frontend/api/pfs';

const useArchiveDownload = (
  projectId: string,
  repoId: string,
  commitOrBranchId: string,
) => {
  const archiveDownload = useCallback(
    async (paths: string[]) => {
      const downloadResponse = await encodeArchiveUrl({
        projectId,
        repoId,
        commitId: commitOrBranchId,
        paths: paths,
      });

      window.open(downloadResponse.url);
    },
    [commitOrBranchId, projectId, repoId],
  );

  return {archiveDownload};
};

export default useArchiveDownload;
