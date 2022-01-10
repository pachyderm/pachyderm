import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection} from '@dash-frontend/lib/types';

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const {isOpen, sidebarSize} = useSidebarInfo();
  const {viewState} = useUrlQueryState();
  const {projectId, jobId} = useUrlState();
  const [dagDirectionSetting] = useLocalProjectSettings({
    projectId,
    key: 'dag_direction',
  });

  const dagDirection =
    viewState.dagDirection || dagDirectionSetting || DagDirection.RIGHT;

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: jobId,
    projectId,
    nodeHeight,
    nodeWidth,
    direction: dagDirection,
  });

  return {
    dags,
    error,
    loading,
    isSidebarOpen: isOpen,
    sidebarSize,
  };
};
