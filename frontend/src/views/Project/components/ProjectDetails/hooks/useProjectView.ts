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
  const {projectId} = useUrlState();
  const [dagDirectionSetting] = useLocalProjectSettings({
    projectId,
    key: 'dag_direction',
  });

  const dagDirection =
    viewState.dagDirection || dagDirectionSetting || DagDirection.RIGHT;

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: viewState.globalIdFilter || undefined,
    projectId,
    nodeHeight,
    nodeWidth,
    direction: dagDirection,
  });

  // list input repos by pipeline, and pipeline outputs by repo
  const inputRepoLinks = useMemo(() => {
    let links = {} as Record<string, string[]>;
    dags?.forEach((dag) => {
      dag.links.forEach((link) => {
        if (!link.target.includes('repo') && !link.target.includes('http')) {
          links = {
            ...links,
            [link.target]: links[link.target]
              ? [...links[link.target], link.source]
              : [link.source],
            [link.source]: links[link.source]
              ? [...links[link.source], link.target]
              : [link.target],
          };
        }
      });
    });
    return links;
  }, [dags]);

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
    inputRepoLinks,
    error,
    loading,
    isSidebarOpen: isOpen,
    sidebarSize,
    repoRedirect,
    pipelineRedirect,
  };
};
