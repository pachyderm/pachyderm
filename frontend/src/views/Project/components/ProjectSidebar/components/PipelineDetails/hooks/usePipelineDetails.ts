import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {LINEAGE_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

const usePipelineDetails = () => {
  const {pipelineId, projectId} = useUrlState();
  const {
    loading: pipelineLoading,
    pipeline,
    isServiceOrSpout,
  } = useCurrentPipeline();
  const {jobs, loading: jobsLoading} = useJobs(
    {
      projectId,
      pipelineId,
      limit: 1,
    },
    {
      skip: !pipeline || pipelineLoading,
    },
  );

  const tabsBasePath = LINEAGE_PIPELINE_PATH;

  return {
    loading: pipelineLoading || jobsLoading,
    pipeline,
    lastJob: jobs[0],
    isServiceOrSpout,
    tabsBasePath,
  };
};

export default usePipelineDetails;
