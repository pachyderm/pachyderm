import {useQuery} from '@tanstack/react-query';

import {listPipeline} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

// This is a really heavy query that is only used for the community edition banner.
// It should not be called when there is a potential for hundreds or thousands of pipelines.
export const useAllPipelines = () => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.allPipelines,
    queryFn: () => listPipeline({details: false}),
  });

  return {
    loading,
    pipelines: data,
    error: getErrorMessage(error),
  };
};
