import usePipeline from './usePipeline';
import useUrlState from './useUrlState';

const useCurrentPipeline = () => {
  const {pipelineId, projectId} = useUrlState();
  const {pipeline, loading} = usePipeline({
    id: pipelineId,
    projectId,
  });

  return {
    pipeline,
    loading,
  };
};

export default useCurrentPipeline;
