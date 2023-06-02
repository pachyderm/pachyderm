import {useCallback} from 'react';

import useFileDownload from '@dash-frontend/hooks/useFileDownload';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  getProxyHostName,
  getTlsEnabled,
} from '@dash-frontend/lib/runtimeVariables';

const useArchiveDownload = () => {
  const {repoId, commitId, branchId, projectId} = useUrlState();

  const {fileDownload} = useFileDownload();

  const archiveDownload = useCallback(
    async (paths: string[]) => {
      const identifier = commitId ? commitId : branchId;

      const downloadResponse = await fileDownload({
        variables: {
          args: {
            projectId,
            repoId,
            commitId: identifier,
            paths: paths,
          },
        },
      });
      //TODO: csrf issue for local dev
      const tls = getTlsEnabled();
      const hostName = getProxyHostName();
      const token = window.localStorage.getItem('auth-token');

      window.open(
        `${tls ? 'https://' : 'http://'}${hostName}${
          downloadResponse.data?.fileDownload
        }${token ? `?authn-token=${token}` : ''}`,
      );
    },
    [branchId, commitId, fileDownload, projectId, repoId],
  );

  return {archiveDownload};
};

export default useArchiveDownload;
