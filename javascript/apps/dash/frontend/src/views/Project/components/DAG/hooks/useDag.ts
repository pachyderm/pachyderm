import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {zoomIdentity} from 'd3-zoom';
import {useEffect} from 'react';

import {Dag, DagDirection} from '@graphqlTypes';
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
  const {selectedNode} = useRouteController();

  // left align dag or vertically align dag
  useEffect(() => {
    const graph = select<SVGGElement, unknown>(`#${id}`);

    const containsSelectedNode = data.nodes.some(
      (n) => n.name === selectedNode,
    );
    const horizontal =
      dagDirection === DagDirection.LEFT || dagDirection === DagDirection.RIGHT;
    const coordinateExtent = horizontal
      ? extent(data.nodes, (d) => d.x)
      : extent(data.nodes, (d) => d.y);
    const min = coordinateExtent[0] || 0;

    const transform = zoomIdentity.translate(
      horizontal && !containsSelectedNode ? -min + nodeWidth : 0,
      !horizontal && !containsSelectedNode ? -min + nodeHeight : 0,
    );

    graph.attr('transform', transform.toString());
  }, [data.nodes, id, nodeHeight, nodeWidth, dagDirection, selectedNode]);
};

export default useDag;
