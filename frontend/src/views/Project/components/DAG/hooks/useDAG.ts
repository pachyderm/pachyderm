import {useMemo} from 'react';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection, InputOutputNodesMap} from '@dash-frontend/lib/types';

import {NODE_HEIGHT, NODE_WIDTH} from './../../../constants/nodeSizes';

export const useDAG = () => {
  const {searchParams} = useUrlQueryState();
  const {projectId} = useUrlState();
  const [dagDirectionSetting] = useLocalProjectSettings({
    projectId,
    key: 'dag_direction',
  });

  const dagDirection = dagDirectionSetting || DagDirection.DOWN;

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: searchParams.globalIdFilter || undefined,
    projectId,
    nodeHeight: NODE_HEIGHT,
    nodeWidth: NODE_WIDTH,
    direction: dagDirection,
  });

  // list pipeline outputs by repo
  const pipelineOutputsMap = useMemo(() => {
    const links: InputOutputNodesMap = {};

    if (!dags) return links;

    for (const dag of dags) {
      for (const link of dag.links) {
        try {
          new URL(link.target.name);
          // ignore any Egress nodes
          continue;
        } catch {
          const sourceList = links[link.source.id] || [];
          sourceList.push(link.target);
          links[link.source.id] = sourceList;
        }
      }
    }
    return links;
  }, [dags]);

  return {
    pipelineOutputsMap,
    dags,
    error,
    loading,
  };
};
