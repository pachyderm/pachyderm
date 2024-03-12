import {alg} from '@dagrejs/graphlib';
import {useMemo} from 'react';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {generateVertexId} from '@dash-frontend/lib/DAG/DAGhelpers';
import {Dags} from '@dash-frontend/lib/types';
import useDAGRouteController from '@dash-frontend/views/Project/components/DAGView/hooks/useDAGRouteController';

const useDag = (data?: Dags) => {
  const {projectId} = useUrlState();
  const {selectedPipeline, selectedRepo, repoPathMatch, pipelinePathMatch} =
    useDAGRouteController();
  const [, translateX, translateY, scale] =
    document
      .getElementById('Dags')
      ?.getAttribute('transform')
      ?.split(/translate\(|\) scale\(|\)|,/)
      ?.map((s) => parseFloat(s)) || [];
  const svgWidth = document.getElementById('Svg')?.clientWidth || 0;
  const svgHeight = document.getElementById('Svg')?.clientHeight || 0;

  const nodeSelected =
    (pipelinePathMatch && selectedPipeline) || (repoPathMatch && selectedRepo);
  const selectedId = nodeSelected && generateVertexId(projectId, nodeSelected);

  const selectedNodePreorder = useMemo(() => {
    if (selectedId && data?.graph) {
      return alg.preorder(data?.graph, [selectedId]);
    }
  }, [data?.graph, selectedId]);

  const selectedNodeReversePreorder = useMemo(() => {
    if (selectedId && data?.reverseGraph) {
      return alg.preorder(data?.reverseGraph, [selectedId]);
    }
  }, [data?.reverseGraph, selectedId]);

  return {
    translateX,
    translateY,
    scale,
    svgWidth,
    svgHeight,
    selectedNodePreorder,
    selectedNodeReversePreorder,
  };
};

export default useDag;
