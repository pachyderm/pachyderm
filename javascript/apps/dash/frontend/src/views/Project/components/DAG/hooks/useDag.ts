import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom, zoomIdentity} from 'd3-zoom';
import {useEffect, useState} from 'react';

import {Dag} from '@graphqlTypes';

import useRouteController from './useRouteController';

type useDagProps = {
  count: number;
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  setLargestDagWidth: React.Dispatch<React.SetStateAction<number | null>>;
  largestDagWidth: number | null;
};

const useDag = ({
  id,
  nodeHeight,
  nodeWidth,
  count,
  largestDagWidth,
  setLargestDagWidth,
  data,
  isInteractive,
}: useDagProps) => {
  const [svgParentSize, setSVGParentSize] = useState({
    width: 0,
    height: 0,
  });
  const {navigateToDag} = useRouteController();

  const parent = select<HTMLTableRowElement, null>(`#${id}Base`);
  const svgElement = select<SVGSVGElement, null>(`#${id}`).node();

  const dagWidth = svgElement ? svgElement.getBBox().width + nodeWidth * 2 : 0;
  const dagHeight = svgElement
    ? svgElement.getBBox().height + nodeHeight * 2
    : 0;

  const parentWidth = parent.node()?.clientWidth || 0;
  const height =
    count > 1 ? dagHeight : Math.max(300, window.innerHeight - 160);

  useEffect(() => {
    const svgElement = select<SVGSVGElement, null>(`#${id}`).node();
    const parentElement = select<HTMLTableRowElement, null>(
      `#${id}Base`,
    ).node();

    if (svgElement && parentElement) {
      const svgWidth = svgElement.getBBox().width + nodeWidth * 2;
      const parentWidth = parentElement.clientWidth;

      // send up DAG width for scaling across DAGs
      if (svgWidth > (largestDagWidth || parentWidth)) {
        setLargestDagWidth(svgWidth);
      }
      // send up DAG size to wrapper
      if (
        svgParentSize.width !== svgElement.getBBox().width ||
        svgParentSize.height !== svgElement.getBBox().height + nodeHeight
      )
        setSVGParentSize({
          width: svgElement.getBBox().width,
          height: svgElement.getBBox().height + nodeHeight,
        });
    }
  }, [
    id,
    setSVGParentSize,
    svgParentSize,
    nodeHeight,
    nodeWidth,
    largestDagWidth,
    setLargestDagWidth,
    data,
  ]);

  useEffect(() => {
    const svg = select<SVGSVGElement, unknown>(`#${id}`);
    const graph = select<SVGGElement, unknown>(`#${id}Graph`);
    const zoomed = (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
      const {transform} = event;
      graph.attr('transform', transform.toString());
    };

    const zoom = d3Zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.6, 1])
      .on('zoom', zoomed);

    svg.call(zoom);

    // initialize zoom based on node positions and center dag
    const xExtent = extent(data.nodes, (d) => d.x);
    const yExtent = extent(data.nodes, (d) => d.y);
    const xMin = xExtent[0] || 0;
    const xMax = xExtent[1] || svgParentSize.width;
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgParentSize.height;

    const transform = zoomIdentity
      .translate(
        svgParentSize.width / 2,
        svgParentSize.height / 2 - nodeHeight / 2,
      )
      .translate(-(xMin + xMax) / 2, -(yMin + yMax) / 2);
    svg.call(zoom.transform, transform);

    !isInteractive && zoom.on('zoom', null);
  }, [data.nodes, id, isInteractive, svgParentSize, nodeHeight]);

  return {
    dagHeight,
    dagWidth,
    height,
    navigateToDag,
    parentWidth,
  };
};

export default useDag;
