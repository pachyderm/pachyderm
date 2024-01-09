import {useQuery, useQueryClient} from '@tanstack/react-query';
import {useCallback} from 'react';

import {Project, ProjectStatus} from '@dash-frontend/api/pfs';
import {PipelineState} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

import {usePipelines} from './usePipelines';

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
  const {pipelines, loading: pipelinesLoading} = usePipelines(projectName);
  const {data, error, isLoading} = useQuery({
    queryKey: queryKeys.projectStatus({projectId: projectName}),
    queryFn: () => {
      if (!pipelines?.length) return ProjectStatus.HEALTHY;

      const status = pipelines.some(
        (pipeline) =>
          pipeline.state === PipelineState.PIPELINE_CRASHING ||
          pipeline.state === PipelineState.PIPELINE_FAILURE,
      )
        ? ProjectStatus.UNHEALTHY
        : ProjectStatus.HEALTHY;

      return status;
    },
    enabled: !!pipelines,
  });

  return {
    error: getErrorMessage(error),
    loading: isLoading || pipelinesLoading,
    projectStatus: data,
  };
};
