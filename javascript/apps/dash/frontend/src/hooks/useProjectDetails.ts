import {useProjectDetailsQuery} from '@dash-frontend/generated/hooks';
import {Project} from '@graphqlTypes';

export const useProjectDetails = (projectId: Project['id']) => {
  const {data, error, loading} = useProjectDetailsQuery({
    variables: {
      args: {
        projectId,
      },
    },
  });

  return {
    error,
    projectDetails: data?.projectDetails,
    loading,
  };
};
