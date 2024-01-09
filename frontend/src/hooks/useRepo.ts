import {useQuery} from '@tanstack/react-query';

import {InspectRepoRequest, inspectRepo} from '@dash-frontend/api/pfs';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useRepo = (req: InspectRepoRequest) => {
  const {data, isLoading, error} = useQuery({
    queryKey: queryKeys.repo({
      projectId: req.repo?.project?.name,
      repoId: req.repo?.name,
    }),
    queryFn: () => inspectRepo(req),
  });

  return {repo: data, loading: isLoading, error};
};
