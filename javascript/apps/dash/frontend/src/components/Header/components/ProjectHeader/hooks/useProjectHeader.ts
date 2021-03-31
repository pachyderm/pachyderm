import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import {JobState} from '@graphqlTypes';

const useProjectHeader = () => {
  const {currentProject, loading} = useCurrentProject();
  const {url} = useRouteMatch();
  const {jobs} = useJobs(currentProject?.id);

  const numOfFailedJobs = useMemo(
    () =>
      jobs.filter(
        (job) => String(JobState[job.state]) === String(JobState.JOB_FAILURE),
      ).length,
    [jobs],
  );

  return {
    projectName: currentProject?.name || '',
    numOfFailedJobs,
    seeJobsUrl: `${url}/jobs`,
    loading,
  };
};

export default useProjectHeader;
