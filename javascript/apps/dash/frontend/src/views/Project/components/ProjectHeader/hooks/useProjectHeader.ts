import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import {JOBS_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  jobsRoute,
  projectRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {JobState} from '@graphqlTypes';

const useProjectHeader = () => {
  const {projectId, currentProject, loading} = useCurrentProject();
  const {jobSets} = useJobSets({projectId});
  const jobsMatch = useRouteMatch({
    path: JOBS_PATH,
    exact: true,
  });

  const numOfFailedJobs = useMemo(
    () => jobSets.filter((job) => job.state === JobState.JOB_FAILURE).length,
    [jobSets],
  );

  const seeJobsUrl = currentProject?.id
    ? jobsRoute({projectId: currentProject?.id})
    : '/';
  const projectUrl = projectRoute({projectId});

  return {
    projectName: currentProject?.name || '',
    numOfFailedJobs,
    seeJobsUrl,
    projectUrl,
    seeJobsOpen: jobsMatch?.isExact,
    loading,
  };
};

export default useProjectHeader;
