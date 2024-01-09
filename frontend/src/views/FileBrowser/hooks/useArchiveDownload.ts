import {useCallback} from 'react';

import {encodeArchiveUrl} from '@dash-frontend/api/pfs';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useArchiveDownload = () => {
  const {repoId, commitId, branchId, projectId} = useUrlState();

  const archiveDownload = useCallback(
    async (paths: string[]) => {
      const identifier = commitId ? commitId : branchId;
      const downloadResponse = await encodeArchiveUrl({
        projectId,
        repoId,
        commitId: identifier,
        paths: paths,
      });

      window.open(downloadResponse.url);
    },
    [branchId, commitId, projectId, repoId],
  );

  return {archiveDownload};
};

export default useArchiveDownload;
