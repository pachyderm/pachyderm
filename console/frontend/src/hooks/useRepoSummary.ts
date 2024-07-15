import {useQuery} from '@tanstack/react-query';

import {reposSummary, Project} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useRepoSummary = (
  projectName: Project['name'],
  enabled = true,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.reposSummary({projectId: projectName}),
    queryFn: () =>
      reposSummary({
        projects: [{name: projectName}],
      }),
    enabled,
  });

  const repoSummary = data?.summaries?.find(
    (summary) => summary.project?.name === projectName,
  );

  return {
    loading,
    repoSummary,
    error: getErrorMessage(error),
  };
};
