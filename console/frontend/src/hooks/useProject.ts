import {useQuery} from '@tanstack/react-query';

import {Project, inspectProject} from '@dash-frontend/api/pfs';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const useProject = (projectName: Project['name']) => {
  const {
    data,
    error,
    isLoading: loading,
  } = useQuery({
    queryKey: queryKeys.project({projectId: projectName}),
    queryFn: () => inspectProject({project: {name: projectName}}),
  });

  return {
    error: getErrorMessage(error),
    loading,
    project: data,
  };
};
