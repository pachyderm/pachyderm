import {PipelineType} from '@graphqlTypes';

import usePipeline from './usePipeline';
import useUrlState from './useUrlState';

const useCurrentPipeline = () => {
  const {pipelineId, projectId} = useUrlState();
  const {pipeline, loading} = usePipeline({
    id: pipelineId,
    projectId,
  });

  const isServiceOrSpout =
    pipeline?.type === PipelineType.SERVICE ||
    pipeline?.type === PipelineType.SPOUT;

  const isSpout = pipeline?.type === PipelineType.SPOUT;

  return {
    pipeline,
    loading,
    isServiceOrSpout,
    isSpout,
    pipelineType: pipeline?.type,
  };
};

export default useCurrentPipeline;
