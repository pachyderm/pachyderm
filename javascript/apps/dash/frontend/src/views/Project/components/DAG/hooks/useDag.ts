import {extent} from 'd3-array';
import {useCallback, useEffect, useState} from 'react';

import {Dag, DagDirection, NodeState} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

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
    const errorNode = data.nodes.find((n) => n.state === NodeState.ERROR);

    if (errorNode) {
      navigateToNode(errorNode);
    } else {
      navigateToNode(data.nodes[0]);
    }
  }, [data.nodes, navigateToNode]);

  // adjust rect for hover state and dag selection
  useEffect(() => {
    const horizontal = dagDirection === DagDirection.RIGHT;
    const yExtent = extent(data.nodes, (d) => d.y);
    const xExtent = extent(data.nodes, (d) => d.x);
    const minY = yExtent[0] || 0;
    const maxY = yExtent[yExtent.length - 1] || 0;
    const minX = xExtent[0] || 0;
    const maxX = xExtent[xExtent.length - 1] || 0;

    setRectBox({
      x: horizontal ? 0 : minX,
      y: horizontal ? minY - nodeHeight / 2 : 0,
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
