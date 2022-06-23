import {useCallback} from 'react';

import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import deriveRepoNameFromNode from '@dash-frontend/lib/deriveRepoNameFromNode';
import {Node} from '@dash-frontend/lib/types';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

export const useProjectDetails = () => {
  const {isOpen, sidebarSize} = useSidebarInfo();
  const {projectId} = useUrlState();

  const repoRedirect = useCallback(
    (node: Node) =>
      repoRoute({
        branchId: 'default',
        projectId,
        repoId: deriveRepoNameFromNode(node),
      }),
    [projectId],
  );

  const pipelineRedirect = useCallback(
    (node: Node) =>
      pipelineRoute({
        projectId,
        pipelineId: node.id,
      }),
    [projectId],
  );

  return {
    isSidebarOpen: isOpen,
    sidebarSize,
    repoRedirect,
    pipelineRedirect,
  };
};
