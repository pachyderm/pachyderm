import {useQuery} from '@tanstack/react-query';

import {inspectRepo} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

import useUrlState from './useUrlState';

export const useCurrentRepo = (enabled = true) => {
  const {repoId, projectId} = useUrlState();
  const {
    data: repo,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.repo({projectId, repoId}),
    queryFn: () =>
      inspectRepo({
        repo: {name: repoId, project: {name: projectId}},
      }),
    enabled,
  });

  return {
    repo,
    error: getErrorMessage(error),
    loading,
  };
};
