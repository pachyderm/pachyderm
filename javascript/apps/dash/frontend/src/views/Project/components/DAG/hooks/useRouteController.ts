import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
  dagRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Dag, Node, NodeType} from '@graphqlTypes';

import deriveRouteParamFromNode from '../utils/deriveRepoNameFromNode';

interface UseRouteControllerArgs {
  dag?: Dag;
}

const useRouteController = ({dag}: UseRouteControllerArgs) => {
  const {projectId, repoId, pipelineId} = useUrlState();

  const browserHistory = useHistory();

  const navigateToDag = useCallback(
    (dagIndex: string) => {
      browserHistory.push(dagRoute({projectId, dagId: dagIndex}));
    },
    [browserHistory, projectId],
  );

  const navigateToNode = useCallback(
    (n: Node, dagId: string) => {
      if (n.type === NodeType.REPO) {
        browserHistory.push(
          repoRoute({projectId, dagId, repoId: deriveRouteParamFromNode(n)}),
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
    [browserHistory, projectId],
  );

  const selectedNode = useMemo(() => {
    if (!pipelineId && !repoId) {
      return undefined;
    }

    if (repoId) {
      return dag?.nodes.find((n) => {
        return (
          n.type === NodeType.REPO && deriveRouteParamFromNode(n) === repoId
        );
      });
    }

    if (pipelineId) {
      return dag?.nodes.find((n) => {
        return n.type === NodeType.PIPELINE && n.name === pipelineId;
      });
    }
  }, [repoId, pipelineId, dag]);

  return {
    selectedNode,
    navigateToNode,
    navigateToDag,
  };
};

export default useRouteController;
