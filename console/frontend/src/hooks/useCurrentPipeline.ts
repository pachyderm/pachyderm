import {useQuery} from '@tanstack/react-query';

import {
  inspectPipeline,
  PipelineInfoPipelineType,
} from '@dash-frontend/api/pps';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

import useUrlState from './useUrlState';

export const useCurrentPipeline = (enabled = true) => {
  const {pipelineId, projectId} = useUrlState();
  const {
    data: pipeline,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipeline({projectId, pipelineId}),
    queryFn: () =>
      inspectPipeline({
        pipeline: {
          project: {
            name: projectId,
          },
          name: pipelineId,
        },
        details: true,
      }),
    enabled,
  });

  const isServiceOrSpout =
    pipeline?.type === PipelineInfoPipelineType.PIPELINE_TYPE_SERVICE ||
    pipeline?.type === PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT;

  const isSpout =
    pipeline?.type === PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT;

  return {
    pipeline,
    loading,
    error: getErrorMessage(error),
    isServiceOrSpout,
    isSpout,
    pipelineType: pipeline?.type,
  };
};
