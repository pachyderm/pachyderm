import {useMemo} from 'react';

import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileUploadRoute} from '@dash-frontend/views/Project/utils/routes';

const useUploadFilesButton = () => {
  const {projectId, repoId} = useUrlState();
  const {loading} = useCurrentRepo();

  const fileUploadPath = useMemo(() => {
    if (projectId && repoId) return fileUploadRoute({projectId, repoId});
  }, [projectId, repoId]);

  return {
    fileUploadPath,
    loading,
  };
};

export default useUploadFilesButton;
