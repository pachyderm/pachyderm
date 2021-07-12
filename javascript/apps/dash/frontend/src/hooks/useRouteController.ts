import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  jobRoute,
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Node, NodeType} from '@graphqlTypes';

import deriveRouteParamFromNode from '../lib/deriveRepoNameFromNode';

import useIsViewingJob from './useIsViewingJob';

const useRouteController = () => {
  const {projectId, repoId, pipelineId, jobId} = useUrlState();
  const isViewingJob = useIsViewingJob();
  const browserHistory = useHistory();

  const navigateToNode = useCallback(
    (n: Node) => {
      if (n.type === NodeType.EGRESS) return;
      else if (
        n.type === NodeType.OUTPUT_REPO ||
        n.type === NodeType.INPUT_REPO
      ) {
        browserHistory.push(
          repoRoute({
            branchId: 'master',
            projectId,
            repoId: deriveRouteParamFromNode(n),
          }),
        );
      } else {
        const pipelineId = deriveRouteParamFromNode(n);

        if (isViewingJob) {
          browserHistory.push(
            jobRoute({
              projectId,
              pipelineId,
              jobId,
            }),
          );
        } else {
          browserHistory.push(
            pipelineRoute({
              projectId,
              pipelineId,
            }),
          );
        }
      }
    },
    [projectId, browserHistory, jobId, isViewingJob],
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
