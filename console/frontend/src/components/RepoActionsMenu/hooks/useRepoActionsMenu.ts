import {useMemo, useCallback} from 'react';
import {useRouteMatch, useHistory} from 'react-router';

import {
  Permission,
  ResourceType,
  useAuthorizeLazy,
} from '@dash-frontend/hooks/useAuthorize';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {PROJECT_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  repoRoute,
  fileUploadRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {
  TrashSVG,
  DropdownItem,
  DirectionsSVG,
  UploadSVG,
} from '@pachyderm/components';

import useDeleteRepoButton from './useDeleteRepoButton';

const useRepoActionsMenu = (repoId: string) => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();

  const {
    getVertices,
    modalOpen: deleteRepoModalOpen,
    setModalOpen: setDeleteModalOpen,
    verticesLoading,
    canDelete,
  } = useDeleteRepoButton(repoId);

  const projectMatch = !!useRouteMatch(PROJECT_PATH);

  const {checkPermissions, hasRepoDelete, hasRepoWrite} = useAuthorizeLazy({
    permissions: [Permission.REPO_DELETE, Permission.REPO_WRITE],
    resource: {type: ResourceType.REPO, name: `${projectId}/${repoId}`},
  });

  const onDropdownMenuSelect = useCallback(
    (id: string) => {
      switch (id) {
        case 'view-dag':
          return browserHistory.push(repoRoute({projectId, repoId}));
        case 'upload-files':
          if (projectId)
            return browserHistory.push(fileUploadRoute({projectId, repoId}));
          return null;
        case 'delete-repo':
          return setDeleteModalOpen(true);
        default:
          return null;
      }
    },
    [browserHistory, projectId, repoId, setDeleteModalOpen],
  );

  const menuItems: DropdownItem[] = useMemo(() => {
    const items: DropdownItem[] = [];

    const deleteTooltipText = () => {
      if (!hasRepoDelete) return 'You need at least repoOwner to delete this.';
      else if (!canDelete)
        return "This repo can't be deleted while it has associated pipelines.";
      return undefined;
    };

    if (projectMatch) {
      items.push({
        id: 'view-dag',
        content: 'View in DAG',
        disabled: false,
        IconSVG: DirectionsSVG,
        closeOnClick: true,
      });
    }
    items.push(
      {
        id: 'upload-files',
        content: 'Upload Files',
        disabled: !hasRepoWrite,
        IconSVG: UploadSVG,
        tooltipText: !hasRepoWrite
          ? 'You need at least repoWriter to upload files.'
          : undefined,
        closeOnClick: true,
      },
      {
        id: 'delete-repo',
        content: 'Delete Repo',
        disabled: verticesLoading || !hasRepoDelete || !canDelete,
        IconSVG: TrashSVG,
        tooltipText: deleteTooltipText(),
        closeOnClick: true,
      },
    );
    return items;
  }, [canDelete, hasRepoDelete, hasRepoWrite, projectMatch, verticesLoading]);

  const onMenuOpen = useCallback(() => {
    getVertices();
    checkPermissions();
  }, [checkPermissions, getVertices]);

  return {
    repoId,
    deleteRepoModalOpen,
    setDeleteModalOpen,
    onDropdownMenuSelect,
    menuItems,
    onMenuOpen,
  };
};

export default useRepoActionsMenu;
