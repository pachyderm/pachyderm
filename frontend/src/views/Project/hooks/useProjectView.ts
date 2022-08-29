import {NodeType} from '@graphqlTypes';
import {useMemo} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection, DagNodes} from '@dash-frontend/lib/types';

import {NODE_HEIGHT, NODE_WIDTH} from './../constants/nodeSizes';

export const useProjectView = () => {
  const {viewState} = useUrlQueryState();
  const {projectId} = useUrlState();
  const [dagDirectionSetting] = useLocalProjectSettings({
    projectId,
    key: 'dag_direction',
  });

  const dagDirection =
    viewState.dagDirection || dagDirectionSetting || DagDirection.DOWN;

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: viewState.globalIdFilter || undefined,
    projectId,
    nodeHeight: NODE_HEIGHT,
    nodeWidth: NODE_WIDTH,
    direction: dagDirection,
  });

  // list input repos by pipeline, and pipeline outputs by repo
  const inputRepoLinks = useMemo(() => {
    let links = {} as Record<string, string[]>;
    dags?.forEach((dag) => {
      dag.links.forEach((link) => {
        if (!link.target.includes('repo')) {
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

  return {
    dags,
    nodes,
    inputRepoLinks,
    error,
    loading,
  };
};
