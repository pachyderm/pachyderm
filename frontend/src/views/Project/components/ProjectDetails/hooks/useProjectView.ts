import {NodeType} from '@graphqlTypes';
import {useCallback, useMemo} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import deriveRepoNameFromNode from '@dash-frontend/lib/deriveRepoNameFromNode';
import {DagDirection, Node, DagNodes} from '@dash-frontend/lib/types';
import {
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const {isOpen, sidebarSize} = useSidebarInfo();
  const {viewState} = useUrlQueryState();
  const {projectId, jobId} = useUrlState();
  const [dagDirectionSetting] = useLocalProjectSettings({
    projectId,
    key: 'dag_direction',
  });

  const dagDirection =
    viewState.dagDirection || dagDirectionSetting || DagDirection.RIGHT;

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: jobId,
    projectId,
    nodeHeight,
    nodeWidth,
    direction: dagDirection,
  });

  // make a list of repos and pipelines per dag
  const nodes = useMemo(
    () =>
      dags
        ? dags.reduce<{repos: DagNodes[]; pipelines: DagNodes[]}>(
            (acc, nextDag) => {
              return nextDag
                ? {
                    repos: [
                      ...acc.repos,
                      {
                        id: nextDag.id,
                        nodes: nextDag.nodes.filter((node) =>
                          [NodeType.INPUT_REPO, NodeType.OUTPUT_REPO].includes(
                            node.type,
                          ),
                        ),
                      },
                    ],
                    pipelines: [
                      ...acc.pipelines,
                      {
                        id: nextDag.id,
                        nodes: nextDag.nodes.filter(
                          (node) => node.type === NodeType.PIPELINE,
                        ),
                      },
                    ],
                  }
                : acc;
            },
            {repos: [], pipelines: []},
          )
        : {repos: [], pipelines: []},
    [dags],
  );

  const repoRedirect = useCallback(
    (node: Node) =>
      repoRoute({
        branchId: 'master',
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
    dags,
    nodes,
    error,
    loading,
    isSidebarOpen: isOpen,
    sidebarSize,
    repoRedirect,
    pipelineRedirect,
  };
};
