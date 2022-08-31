import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Node} from '@dash-frontend/lib/types';
import {
  LINEAGE_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  jobRoute,
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

import deriveRouteParamFromNode from '../lib/deriveRepoNameFromNode';

import useIsViewingJob from './useIsViewingJob';
const useRouteController = () => {
  const {projectId, repoId, pipelineId, jobId} = useUrlState();
  const isViewingJob = useIsViewingJob();
  const browserHistory = useHistory();

  const navigateToNode = useCallback(
    (n: Node, destination: 'pipeline' | 'repo') => {
      if (destination === 'repo') {
        browserHistory.push(
          repoRoute({
            branchId: 'default',
            projectId,
            repoId: deriveRouteParamFromNode(n),
          }),
        );
      } else if (destination === 'pipeline') {
        const pipelineId = n.id;

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

  const pipelinePathMatch = useRouteMatch({
    path: LINEAGE_PIPELINE_PATH,
  });

  const repoPathMatch = useRouteMatch({
    path: LINEAGE_REPO_PATH,
  });

  return {
    selectedRepo: repoId,
    selectedPipeline: pipelineId,
    selectedNode: pipelineId || repoId,
    navigateToNode,
    repoPathMatch,
    pipelinePathMatch,
  };
};

export default useRouteController;
