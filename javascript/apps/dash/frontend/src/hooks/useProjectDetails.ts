import {useProjectDetailsQuery} from '@dash-frontend/generated/hooks';
import {Project} from '@graphqlTypes';

export const useProjectDetails = (
  projectId: Project['id'],
  jobsLimit?: number,
) => {
  const {data, error, loading} = useProjectDetailsQuery({
    variables: {
      args: {
        projectId,
        jobsLimit,
      },
    },
  });

  return {
    error,
    projectDetails: data?.projectDetails,
    loading,
  };
};
