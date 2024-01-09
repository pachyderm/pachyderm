import {JobArgs, useJob} from './useJob';
import {useJobs} from './useJobs';

export const useJobOrJobs = (
  {id, pipelineName, projectId}: JobArgs,
  enabled = true,
) => {
  const {
    job: inspectJobData,
    loading: inspectJobLoading,
    error: inspectJobError,
  } = useJob(
    {
      id,
      pipelineName: pipelineName,
      projectId,
    },
    enabled && !!id,
  );

  const {
    jobs: listJobData,
    loading: listJobLoading,
    error: listJobError,
  } = useJobs(
    {
      projectName: projectId,
      pipelineIds: pipelineName ? [pipelineName] : undefined,
      limit: 1,
    },
    enabled && !id,
  );

  return {
    job: id ? inspectJobData : listJobData?.[0],
    loading: id ? inspectJobLoading : listJobLoading,
    error: id ? inspectJobError : listJobError,
  };
};
