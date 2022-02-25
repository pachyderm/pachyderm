import {useRouteMatch} from 'react-router';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {
  LINEAGE_PATH,
  LINEAGE_PIPELINE_PATH,
  PROJECT_PIPELINE_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

import {TAB_ID, TAB_IDS} from '../constants/tabIds';

const usePipelineDetails = () => {
  const {loading, pipeline, isServiceOrSpout} = useCurrentPipeline();

  const filteredTabIds = TAB_IDS.filter(
    (tabId) => tabId !== TAB_ID.JOBS || (!loading && !isServiceOrSpout),
  );

  const lineageMatch = useRouteMatch({
    path: LINEAGE_PATH,
  });

  const tabsBasePath = lineageMatch
    ? LINEAGE_PIPELINE_PATH
    : PROJECT_PIPELINE_PATH;

  return {
    loading,
    pipelineName: pipeline?.name,
    filteredTabIds,
    isServiceOrSpout,
    tabsBasePath,
  };
};

export default usePipelineDetails;
