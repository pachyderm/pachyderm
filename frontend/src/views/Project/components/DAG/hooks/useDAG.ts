import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection} from '@dash-frontend/lib/types';

import {NODE_HEIGHT, NODE_WIDTH} from './../../../constants/nodeSizes';

export const useDAG = () => {
  const {viewState} = useUrlQueryState();
  const {projectId} = useUrlState();
  const [dagDirectionSetting] = useLocalProjectSettings({
    projectId,
    key: 'dag_direction',
  });

  const dagDirection =
    viewState.dagDirection || dagDirectionSetting || DagDirection.DOWN;

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: viewState.globalIdFilter || undefined,
    projectId,
    nodeHeight: NODE_HEIGHT,
    nodeWidth: NODE_WIDTH,
    direction: dagDirection,
  });

  return {
    dags,
    error,
    loading,
  };
};
