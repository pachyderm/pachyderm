import {QueryFunctionOptions} from '@apollo/client';
import {JobSetsQueryArgs} from '@graphqlTypes';

import {JOBS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useJobSetsQuery} from '@dash-frontend/generated/hooks';

export const useJobSets = (
  args: JobSetsQueryArgs,
  opts?: QueryFunctionOptions,
) => {
  const {data, error, loading} = useJobSetsQuery({
    variables: {args},
    pollInterval: JOBS_POLL_INTERVAL_MS,
    skip: opts?.skip,
  });

  return {
    error,
    jobSets: data?.jobSets || [],
    loading,
  };
};
