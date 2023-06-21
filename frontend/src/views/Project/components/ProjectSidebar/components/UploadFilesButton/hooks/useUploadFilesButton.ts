import {useMemo} from 'react';

import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {hasAtLeastRole} from '@dash-frontend/lib/rbac';
import {fileUploadRoute} from '@dash-frontend/views/Project/utils/routes';

const useUploadFilesButton = () => {
  const {projectId, repoId} = useUrlState();
  const {repo, loading} = useCurrentRepo();

  const hasAuthUploadRepo = hasAtLeastRole(
    'repoWriter',
    repo?.authInfo?.rolesList,
  );

  const tooltipText = !hasAuthUploadRepo
    ? 'You need at least repoWriter to upload files.'
    : 'Upload files';

  const fileUploadPath = useMemo(() => {
    if (projectId && repoId) return fileUploadRoute({projectId, repoId});
  }, [projectId, repoId]);

  const disableButton = loading || !fileUploadPath || !hasAuthUploadRepo;

  return {
    fileUploadPath,
    tooltipText,
    disableButton,
  };
};

export default useUploadFilesButton;
