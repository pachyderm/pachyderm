import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import usePipeline from '@dash-frontend/hooks/usePipeline';
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
  const {pipeline, loading} = usePipeline({
    id: pipelineId,
    projectId,
  });

  const handleSwitch = useCallback(
    (id: string) => {
      if (id === TAB_ID.INFO) {
        browserHistory.push(pipelineRoute({projectId, pipelineId}));
      } else {
        browserHistory.push(
          pipelineRoute({projectId, pipelineId, tabId: TAB_ID.JOBS}),
        );
      }
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
