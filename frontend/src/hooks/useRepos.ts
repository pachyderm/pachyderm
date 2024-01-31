import {useQuery} from '@tanstack/react-query';

import {listRepo, Project} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useRepos = (
  projectName: Project['name'],
  enabled = true,
  staleTime?: number,
) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.repos({projectId: projectName}),
    queryFn: () => listRepo({projects: [{name: projectName}], type: 'user'}),
    enabled,
    staleTime,
  });

  return {
    loading,
    repos: data,
    error: getErrorMessage(error),
  };
};

export const useReposLazy = (projectName: Project['name']) => {
  const {
    data,
    isLoading: loading,
    isFetched,
    error,
    refetch,
  } = useQuery({
    queryKey: queryKeys.repos({projectId: projectName}),
    queryFn: () => listRepo({projects: [{name: projectName}], type: 'user'}),
    enabled: false,
  });

  return {
    getRepos: refetch,
    loading,
    repos: data,
    isFetched,
    error: getErrorMessage(error),
  };
};
