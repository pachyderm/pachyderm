import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {logsViewerPipelneRoute} from '@dash-frontend/views/Project/utils/routes';

import {TAB_ID, TAB_IDS} from '../constants/tabIds';

const usePipelineDetails = () => {
  const {pipelineId, projectId} = useUrlState();
  const {loading, pipeline, isServiceOrSpout} = useCurrentPipeline();

  const filteredTabIds = TAB_IDS.filter(
    (tabId) => tabId !== TAB_ID.JOBS || (!loading && !isServiceOrSpout),
  );

  const pipelineLogsRoute = logsViewerPipelneRoute({
    projectId,
    pipelineId: pipelineId,
  });

  return {
    loading,
    pipelineName: pipeline?.name,
    filteredTabIds,
    isServiceOrSpout,
    pipelineLogsRoute,
  };
};

export default usePipelineDetails;
