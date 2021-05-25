import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {zoomIdentity} from 'd3-zoom';
import {useCallback, useEffect, useState} from 'react';

import {Dag, DagDirection} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

import convertNodeStateToDagState from '../utils/convertNodeStateToDagState';

type useDagProps = {
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  dagDirection: DagDirection;
};

const useDag = ({
  id,
  nodeHeight,
  nodeWidth,
  data,
  dagDirection,
}: useDagProps) => {
  const {selectedNode, navigateToNode} = useRouteController();
  const [rectBox, setRectBox] = useState({x: 0, y: 0, width: 0, height: 0});

  const handleRectClick = useCallback(() => {
    const errorNode = data.nodes.find(
      (n) => convertNodeStateToDagState(n.state) === 'error',
    );

    if (errorNode) {
      navigateToNode(errorNode);
    } else {
      navigateToNode(data.nodes[0]);
    }
  }, [data.nodes, navigateToNode]);

  // left align dag or vertically align dag
  useEffect(() => {
    const graph = select<SVGGElement, unknown>(`#${id}`);

    const containsSelectedNode = data.nodes.some(
      (n) => n.name === selectedNode,
    );
    const horizontal =
      dagDirection === DagDirection.LEFT || dagDirection === DagDirection.RIGHT;
    const yExtent = extent(data.nodes, (d) => d.y);
    const xExtent = extent(data.nodes, (d) => d.x);
    const minY = yExtent[0] || 0;
    const maxY = yExtent[yExtent.length - 1] || 0;
    const minX = xExtent[0] || 0;
    const maxX = xExtent[xExtent.length - 1] || 0;

    const transform = zoomIdentity.translate(
      horizontal && !containsSelectedNode ? -minX + nodeWidth : 0,
      !horizontal && !containsSelectedNode ? -minY + nodeHeight : 0,
    );

    graph.attr('transform', transform.toString());

    // adjust rect for hover state and dag selection
    setRectBox({
      x: minX,
      y: minY - nodeHeight / 2,
      width: maxX - minX + nodeWidth,
      height: maxY - minY + nodeHeight * 1.5,
    });
  }, [
    data.nodes,
    id,
    nodeHeight,
    nodeWidth,
    dagDirection,
    selectedNode,
    setRectBox,
  ]);

  return {handleRectClick, rectBox};
};

export default useDag;
