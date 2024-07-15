import {useQuery} from '@tanstack/react-query';

import {JobSet, inspectJobSet} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useJobSet = (jobSetId: JobSet['id'], enabled = true) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.jobSet({jobSetId: jobSetId}),
    queryFn: () =>
      inspectJobSet({
        jobSet: {
          id: jobSetId,
        },
        details: true,
      }),
    enabled,
  });

  return {
    loading,
    jobSet: data,
    error: getErrorMessage(error),
  };
};

export const useJobSetLazy = (jobSetId: JobSet['id']) => {
  const {
    data,
    isLoading: loading,
    error,
    refetch,
    isFetched,
    isRefetching,
    isSuccess,
  } = useQuery({
    queryKey: queryKeys.jobSet({jobSetId: jobSetId}),
    queryFn: () =>
      inspectJobSet({
        jobSet: {
          id: jobSetId,
        },
        details: true,
      }),
    enabled: false,
  });

  return {
    refetch,
    isFetched,
    isRefetching,
    isSuccess,
    loading,
    jobSet: data,
    error: getErrorMessage(error),
  };
};
