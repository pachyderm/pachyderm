import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Node} from '@dash-frontend/lib/types';
import {
  LINEAGE_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

const useRouteController = () => {
  const {projectId, repoId, pipelineId} = useUrlState();
  const browserHistory = useHistory();

  const navigateToNode = useCallback(
    (n: Node, destination: 'pipeline' | 'repo') => {
      if (destination === 'repo') {
        browserHistory.push(
          repoRoute({
            projectId,
            repoId: n.name,
          }),
        );
      } else if (destination === 'pipeline') {
        browserHistory.push(
          pipelineRoute({
            projectId,
            pipelineId: n.name,
          }),
        );
      }
    },
    [projectId, browserHistory],
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
