import {JOBS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useJobsQuery} from '@dash-frontend/generated/hooks';
import {JobsQueryArgs} from '@graphqlTypes';

export const useJobs = (args: JobsQueryArgs) => {
  const {data, error, loading} = useJobsQuery({
    variables: {args},
    pollInterval: JOBS_POLL_INTERVAL_MS,
  });

  return {
    error,
    jobs: data?.jobs || [],
    loading,
  };
};
