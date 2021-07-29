import {JobSetQueryArgs} from '@graphqlTypes';

import {JOBS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useJobSetQuery} from '@dash-frontend/generated/hooks';

export const useJobSet = (args: JobSetQueryArgs) => {
  const {data, error, loading} = useJobSetQuery({
    variables: {args},
    pollInterval: JOBS_POLL_INTERVAL_MS,
  });

  return {
    error,
    jobSet: data?.jobSet,
    loading,
  };
};
