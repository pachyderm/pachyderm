import {useMemo} from 'react';

import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {JobState} from '@graphqlTypes';

const useProjectHeader = () => {
  const {projectId, currentProject, loading} = useCurrentProject();
  const {jobs} = useJobs({projectId});

  const numOfFailedJobs = useMemo(
    () => jobs.filter((job) => job.state === JobState.JOB_FAILURE).length,
    [jobs],
  );

  return {
    projectName: currentProject?.name || '',
    numOfFailedJobs,
    seeJobsUrl: currentProject?.id
      ? jobsRoute({projectId: currentProject?.id})
      : '/',
    loading,
  };
};

export default useProjectHeader;
