import {useQuery} from '@tanstack/react-query';

import {
  getProjectDefaults,
  GetProjectDefaultsRequest,
} from '@dash-frontend/api/pps';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useGetProjectDefaults = (req: GetProjectDefaultsRequest) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.projectDefaults({projectId: req.project?.name || ''}),
    queryFn: () => getProjectDefaults(req),
  });

  return {
    error,
    projectDefaults: data,
    loading,
  };
};
