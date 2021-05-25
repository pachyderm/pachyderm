import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Node, NodeType} from '@graphqlTypes';

import deriveRouteParamFromNode from '../lib/deriveRepoNameFromNode';

const useRouteController = () => {
  const {projectId, repoId, pipelineId} = useUrlState();

  const browserHistory = useHistory();

  const navigateToNode = useCallback(
    (n: Node) => {
      if (n.type === NodeType.EGRESS) return;
      else if (n.type === NodeType.REPO) {
        browserHistory.push(
          repoRoute({
            branchId: 'master',
            projectId,
            repoId: deriveRouteParamFromNode(n),
          }),
        );
      } else {
        browserHistory.push(
          pipelineRoute({
            projectId,
            pipelineId: deriveRouteParamFromNode(n),
          }),
        );
      }
    },
    [browserHistory, projectId],
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
  };
};

export default useRouteController;
