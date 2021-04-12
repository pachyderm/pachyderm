import {useParams} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const {projectId} = useParams<{projectId: string}>();
  const {dags, loading, error} = useProjectDagsData({
    projectId,
    nodeHeight,
    nodeWidth,
  });

  return {
    dagCount: dags?.length || 0,
    dags,
    error,
    loading,
  };
};
