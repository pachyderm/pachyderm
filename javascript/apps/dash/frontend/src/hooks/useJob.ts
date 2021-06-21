import {useJobQuery} from '@dash-frontend/generated/hooks';
import {JobQueryArgs} from '@graphqlTypes';
import {JOBS_POLL_INTERVAL_MS} from 'constants/pollIntervals';

export const useJob = (args: JobQueryArgs) => {
  const {data, error, loading} = useJobQuery({
    variables: {args},
    pollInterval: JOBS_POLL_INTERVAL_MS,
  });

  return {
    error,
    job: data?.job,
    loading,
  };
};
