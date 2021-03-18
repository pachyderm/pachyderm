import {useParams} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';

export const useProjectView = () => {
  const {projectId} = useParams<{projectId: string}>();
  const {dags, loading, error} = useProjectDagsData(projectId);

  return {
    dagCount: dags?.length || 0,
    dags,
    error,
    loading,
  };
};
