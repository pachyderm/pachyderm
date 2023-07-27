import {Project} from '@graphqlTypes';

import {PROJECTS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useProjectStatusQuery} from '@dash-frontend/generated/hooks';

export const useProjectStatus = (projectId: Project['id']) => {
  const {data, ...rest} = useProjectStatusQuery({
    variables: {id: projectId},
    pollInterval: PROJECTS_POLL_INTERVAL_MS,
  });

  return {
    ...rest,
    projectStatus: data?.projectStatus?.status,
  };
};
