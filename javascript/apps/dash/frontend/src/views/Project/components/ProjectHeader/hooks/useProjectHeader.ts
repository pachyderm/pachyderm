import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';
import {usePipelineJobs} from '@dash-frontend/hooks/usePipelineJobs';
import {JOBS_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {PipelineJobState} from '@graphqlTypes';

const useProjectHeader = () => {
  const {projectId, currentProject, loading} = useCurrentProject();
  const {pipelineJobs} = usePipelineJobs({projectId});
  const jobsMatch = useRouteMatch({
    path: JOBS_PATH,
    exact: true,
  });

  const numOfFailedJobs = useMemo(
    () =>
      pipelineJobs.filter((job) => job.state === PipelineJobState.JOB_FAILURE)
        .length,
    [pipelineJobs],
  );

  return {
    projectName: currentProject?.name || '',
    numOfFailedJobs,
    seeJobsUrl: currentProject?.id
      ? jobsRoute({projectId: currentProject?.id})
      : '/',
    seeJobsOpen: jobsMatch?.isExact,
    loading,
  };
};

export default useProjectHeader;
