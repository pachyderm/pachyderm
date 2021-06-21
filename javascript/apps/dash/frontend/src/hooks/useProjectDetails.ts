import {PROJECTS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
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
    pollInterval: PROJECTS_POLL_INTERVAL_MS,
  });

  return {
    error,
    projectDetails: data?.projectDetails,
    loading,
  };
};
