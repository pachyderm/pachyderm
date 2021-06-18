import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';

import {TAB_ID, TAB_IDS} from '../constants/tabIds';

const usePipelineDetails = () => {
  const {loading, pipeline, isServiceOrSpout} = useCurrentPipeline();

  const filteredTabIds = TAB_IDS.filter(
    (tabId) => tabId !== TAB_ID.JOBS || (!loading && !isServiceOrSpout),
  );

  return {
    loading,
    pipelineName: pipeline?.name,
    filteredTabIds,
    isServiceOrSpout,
  };
};

export default usePipelineDetails;
