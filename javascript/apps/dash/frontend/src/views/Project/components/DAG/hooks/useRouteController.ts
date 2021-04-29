import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
  dagRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Node, NodeType} from '@graphqlTypes';

import deriveRouteParamFromNode from '../utils/deriveRepoNameFromNode';

const useRouteController = () => {
  const {projectId, repoId, pipelineId, dagId} = useUrlState();

  const browserHistory = useHistory();

  const navigateToDag = useCallback(
    (dagIndex: string) => {
      browserHistory.push(dagRoute({projectId, dagId: dagIndex}));
    },
    [browserHistory, projectId],
  );

  const navigateToNode = useCallback(
    (n: Node) => {
      if (n.type === NodeType.REPO) {
        browserHistory.push(
          repoRoute({
            branchId: 'master',
            projectId,
            dagId,
            repoId: deriveRouteParamFromNode(n),
          }),
        );
      } else {
        browserHistory.push(
          pipelineRoute({
            projectId,
            dagId,
            pipelineId: deriveRouteParamFromNode(n),
          }),
        );
      }
    },
    [browserHistory, projectId, dagId],
  );

  const selectedNode = useMemo(() => {
    if (!pipelineId && !repoId) {
      return undefined;
    }

    if (repoId) {
      return `${repoId}_repo`;
    }

    if (pipelineId) {
      return pipelineId;
    }
  }, [repoId, pipelineId]);

  return {
    selectedNode,
    navigateToNode,
    navigateToDag,
  };
};

export default useRouteController;
