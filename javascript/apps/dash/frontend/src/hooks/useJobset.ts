import {JOBS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useJobsetQuery} from '@dash-frontend/generated/hooks';
import {JobsetQueryArgs} from '@graphqlTypes';

export const useJobset = (args: JobsetQueryArgs) => {
  const {data, error, loading} = useJobsetQuery({
    variables: {args},
    pollInterval: JOBS_POLL_INTERVAL_MS,
  });

  return {
    error,
    jobset: data?.jobset,
    loading,
  };
};
