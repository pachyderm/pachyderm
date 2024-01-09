import {useQuery} from '@tanstack/react-query';

import {ListDatumRequest, listDatumPaged} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useDatumsPaged = (req: ListDatumRequest, enabled = true) => {
  const {
    data: datumsPaged,
    refetch,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: [
      ...queryKeys.datums({
        projectId: req.job?.pipeline?.project?.name,
        pipelineId: req.job?.pipeline?.name,
        jobId: req.job?.id,
        args: req,
      }),
    ],
    queryFn: () => listDatumPaged(req),
    enabled,
    refetchInterval: false,
  });

  return {
    ...datumsPaged,
    loading,
    error: getErrorMessage(error),
    refetch,
  };
};
