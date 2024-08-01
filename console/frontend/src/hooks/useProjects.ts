import {useQuery} from '@tanstack/react-query';

import {listProject} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useProjects = () => {
  const {
    data,
    error,
    isLoading: loading,
  } = useQuery({
    queryKey: queryKeys.projects,
    queryFn: () => listProject(),
  });

  return {
    error: getErrorMessage(error),
    loading,
    projects: data ?? [],
  };
};
