import {useParams, useRouteMatch} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';

export const useProjectView = () => {
  const {projectId} = useParams<{projectId: string}>();
  const {path} = useRouteMatch();
  const {dags, loading, error} = useProjectDagsData(projectId);

  return {
    dagCount: dags?.length || 0,
    path,
    dags,
    error,
    loading,
  };
};
