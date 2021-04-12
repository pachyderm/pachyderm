import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Dag, Node, NodeType} from '@graphqlTypes';

import deriveRouteParamFromNode from '../utils/deriveRepoNameFromNode';

interface UseRouteControllerArgs {
  dag: Dag;
}

const useRouteController = ({dag}: UseRouteControllerArgs) => {
  const {projectId, repoId, pipelineId} = useUrlState();

  const browserHistory = useHistory();

  const navigateToNode = useCallback(
    (n: Node) => {
      if (n.type === NodeType.Repo) {
        browserHistory.push(
          repoRoute({projectId, repoId: deriveRouteParamFromNode(n)}),
        );
      } else {
        browserHistory.push(
          pipelineRoute({projectId, pipelineId: deriveRouteParamFromNode(n)}),
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
      return dag.nodes.find((n) => {
        return (
          n.type === NodeType.Repo && deriveRouteParamFromNode(n) === repoId
        );
      });
    }

    if (pipelineId) {
      return dag.nodes.find((n) => {
        return n.type === NodeType.Pipeline && n.name === pipelineId;
      });
    }
  }, [repoId, pipelineId, dag]);

  return {
    selectedNode,
    navigateToNode,
  };
};

export default useRouteController;
