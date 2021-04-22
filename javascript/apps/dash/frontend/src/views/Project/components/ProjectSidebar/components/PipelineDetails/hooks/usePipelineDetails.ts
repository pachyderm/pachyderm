import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';

const usePipelineDetails = () => {
  const {loading, pipeline} = useCurrentPipeline();

  return {
    loading,
    pipelineName: pipeline?.name,
  };
};

export default usePipelineDetails;
