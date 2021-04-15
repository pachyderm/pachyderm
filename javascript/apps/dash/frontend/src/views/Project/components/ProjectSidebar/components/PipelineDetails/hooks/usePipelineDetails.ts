import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

import {TAB_ID, TAB_IDS} from '../constants/tabIds';

const deriveInitialTabId = (tabId: string) => {
  return TAB_IDS.includes(tabId) ? tabId : TAB_ID.INFO;
};

const usePipelineDetails = () => {
  const {params} = useRouteMatch<{tabId?: string}>();
  const browserHistory = useHistory();
  const {pipelineId, projectId} = useUrlState();

  const initialActiveTabId = deriveInitialTabId(params.tabId || '');
  const {loading, pipeline} = useCurrentPipeline();

  const handleSwitch = useCallback(
    (tabId: string) => {
      browserHistory.push(pipelineRoute({projectId, pipelineId, tabId}));
    },
    [browserHistory, pipelineId, projectId],
  );

  return {
    initialActiveTabId,
    loading,
    pipelineName: pipeline?.name,
    handleSwitch,
  };
};

export default usePipelineDetails;
