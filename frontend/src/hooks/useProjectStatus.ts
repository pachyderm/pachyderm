import {useQuery, useQueryClient} from '@tanstack/react-query';
import {useCallback} from 'react';

import {Project, ProjectStatus} from '@dash-frontend/api/pfs';
import {JobState, PipelineInfo, listPipeline} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';
import {NodeState} from '@dash-frontend/lib/types';

/*
We want to filter out unrunnable jobs because they are a result 
of upstream subjob failures. This is only really an issue if a 
user does not have access to the entire DAG and can not see the 
upstream failures that caused the job to become unrunnable.
*/
export const deriveProjectStatus = (pipelines?: PipelineInfo[]) => {
  if (!pipelines?.length) return ProjectStatus.HEALTHY;

  const status = pipelines.some(
    (pipeline) =>
      restPipelineStateToNodeState(pipeline.state) === NodeState.ERROR ||
      (restJobStateToNodeState(pipeline.lastJobState) === NodeState.ERROR &&
        pipeline.lastJobState !== JobState.JOB_UNRUNNABLE),
  )
    ? ProjectStatus.UNHEALTHY
    : ProjectStatus.HEALTHY;

  return status;
};

export const useGetProjectStatus = () => {
  const client = useQueryClient();
  const getProjectStatus = useCallback(
    (projectName: Project['name']) => {
      const result = client.getQueryData<ProjectStatus>(
        queryKeys.projectStatus({projectId: projectName}),
      );

      return result;
    },
    [client],
  );

  return getProjectStatus;
};

export const useProjectStatus = (projectName: Project['name']) => {
  const {data, error, isLoading} = useQuery({
    queryKey: queryKeys.projectStatus({projectId: projectName}),
    queryFn: async () => {
      const pipelines = await listPipeline({projects: [{name: projectName}]});
      return deriveProjectStatus(pipelines);
    },
  });

  return {
    error: getErrorMessage(error),
    loading: isLoading,
    projectStatus: data,
  };
};
