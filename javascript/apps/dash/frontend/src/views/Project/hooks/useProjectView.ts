import {useParams, useRouteMatch} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const {projectId} = useParams<{projectId: string}>();
  const {path} = useRouteMatch();
  const {dags, loading, error} = useProjectDagsData({
    projectId,
    nodeHeight,
    nodeWidth,
  });

  return {
    dagCount: dags?.length || 0,
    path,
    dags,
    error,
    loading,
  };
};
