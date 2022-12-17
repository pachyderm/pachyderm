import {QueryFunctionOptions} from '@apollo/client';
import {JobQueryArgs} from '@graphqlTypes';

import {useJobQuery} from '@dash-frontend/generated/hooks';
import {JOBS_POLL_INTERVAL_MS} from 'constants/pollIntervals';

export const useJob = (args: JobQueryArgs, opts?: QueryFunctionOptions) => {
  const {data, error, loading} = useJobQuery({
    variables: {args},
    pollInterval: JOBS_POLL_INTERVAL_MS,
    skip: opts?.skip,
  });

  return {
    error,
    job: data?.job,
    loading,
  };
};
