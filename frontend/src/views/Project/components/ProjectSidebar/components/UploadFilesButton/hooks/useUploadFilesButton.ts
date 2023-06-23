import {Permission, ResourceType} from '@graphqlTypes';
import {useMemo} from 'react';

import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';
import {fileUploadRoute} from '@dash-frontend/views/Project/utils/routes';

const useUploadFilesButton = () => {
  const {projectId, repoId} = useUrlState();
  const {repo, loading} = useCurrentRepo();

  const {isAuthorizedAction: hasAuthUploadRepo} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_WRITE],
    resource: {type: ResourceType.REPO, name: `${projectId}/${repo?.id}`},
  });

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
