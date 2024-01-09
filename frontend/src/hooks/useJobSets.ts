import {useQuery} from '@tanstack/react-query';

import {Project} from '@dash-frontend/api/pfs';
import {listJobSets} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useJobSets = (projectName: Project['name'], limit = 10) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.jobSets({projectId: projectName, args: {limit}}),
    queryFn: () =>
      listJobSets(
        {
          projects: [{name: projectName}],
          history: '-1',
          details: false,
        },
        limit,
      ),
  });

  return {
    loading,
    jobs: data,
    error: getErrorMessage(error),
  };
};
