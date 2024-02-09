import {useQuery} from '@tanstack/react-query';

import {Project} from '@dash-frontend/api/pfs';
import {listJobSets} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useJobSets = (
  {
    projectName,
    jqFilter,
  }: {
    projectName: Project['name'];
    jqFilter?: string;
  },
  limit = 10,
  enabled = true,
) => {
  const {data, isLoading, error} = useQuery({
    queryKey: queryKeys.jobSets({
      projectId: projectName,
      args: {limit, jqFilter: jqFilter || ''},
    }),
    queryFn: () =>
      listJobSets(
        {
          projects: [{name: projectName}],
          history: '-1',
          details: false,
          jqFilter,
        },
        limit,
      ),
    enabled,
  });

  return {
    loading: isLoading,
    jobs: data,
    error: getErrorMessage(error),
  };
};
