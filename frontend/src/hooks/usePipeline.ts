import {useQuery} from '@tanstack/react-query';

import {
  InspectPipelineRequest,
  inspectPipeline,
  PipelineInfoPipelineType,
} from '@dash-frontend/api/pps';
import {isNotFound} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

export const usePipeline = (req: InspectPipelineRequest, enabled = true) => {
  const {
    data: pipeline,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipeline({
      projectId: req.pipeline?.project?.name,
      pipelineId: req.pipeline?.name,
    }),
    throwOnError: (e) => !isNotFound(e),
    queryFn: () => inspectPipeline(req),
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
  };
};

export const usePipelineLazy = (req: InspectPipelineRequest) => {
  const {
    refetch,
    data: pipeline,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.pipeline({
      projectId: req.pipeline?.project?.name,
      pipelineId: req.pipeline?.name,
    }),
    queryFn: () => inspectPipeline(req),
    enabled: false,
  });

  const isServiceOrSpout =
    pipeline?.type === PipelineInfoPipelineType.PIPELINE_TYPE_SERVICE ||
    pipeline?.type === PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT;

  const isSpout =
    pipeline?.type === PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT;

  return {
    getPipeline: refetch,
    pipeline,
    loading,
    error: getErrorMessage(error),
    isServiceOrSpout,
    isSpout,
  };
};
