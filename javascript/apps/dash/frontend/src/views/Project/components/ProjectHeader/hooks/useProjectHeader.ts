import {useMemo} from 'react';

import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';
import {usePipelineJobs} from '@dash-frontend/hooks/usePipelineJobs';
import {jobsRoute} from '@dash-frontend/views/Project/utils/routes';
import {PipelineJobState} from '@graphqlTypes';

const useProjectHeader = () => {
  const {projectId, currentProject, loading} = useCurrentProject();
  const {pipelineJobs} = usePipelineJobs({projectId});

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
    loading,
  };
};

export default useProjectHeader;
