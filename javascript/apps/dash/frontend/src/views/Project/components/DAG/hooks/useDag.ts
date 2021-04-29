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
  setLargestDagScale: React.Dispatch<React.SetStateAction<number | null>>;
  largestDagScale: number | null;
};

const useDag = ({
  id,
  nodeHeight,
  nodeWidth,
  largestDagScale,
  setLargestDagScale,
  data,
  isInteractive,
}: useDagProps) => {
  const {navigateToDag} = useRouteController();
  const [svgSize, setSVGSize] = useState({
    width: 0,
    height: 0,
  });

  // adjust dag size on data and view changes
  useEffect(() => {
    const parent = select<HTMLTableRowElement, null>(`#${id}Base`);
    const parentWidth = parent.node()?.clientWidth || 0;

    const svgElement = select<SVGSVGElement, null>(`#${id}`).node();
    const dagWidth = svgElement
      ? svgElement.getBBox().width + nodeWidth * 2
      : 0;

    if (isInteractive) {
      setSVGSize({
        height: Math.max(300, window.innerHeight - 140),
        width: parentWidth,
      });
    } else {
      const dagHeight = svgElement
        ? svgElement.getBBox().height + nodeHeight * 2
        : 0;

      setSVGSize({
        height: dagHeight,
        width: parentWidth,
      });

      const scale = Math.min(1, parentWidth / dagWidth);

      if (!largestDagScale) setLargestDagScale(scale);
      else if (scale < largestDagScale) {
        setLargestDagScale(scale);
      }
    }
  }, [
    isInteractive,
    id,
    nodeHeight,
    data,
    largestDagScale,
    setLargestDagScale,
    nodeWidth,
  ]);

  // left align dag, vertically center, scale based on the largest dag, and enable zooming for live dags
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

    const yExtent = extent(data.nodes, (d) => d.y);
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgSize.height;
    const scale = (!isInteractive && largestDagScale) || 1;

    const transform = zoomIdentity
      .translate((nodeWidth / 2) * scale, svgSize.height / 2 - 32) // 32 is the top padding of ${id}Base
      .translate(0, -(yMin + yMax) / 2)
      .scale(scale);
    svg.call(zoom.transform, transform);

    !isInteractive && zoom.on('zoom', null);
  }, [
    data.nodes,
    id,
    isInteractive,
    nodeHeight,
    svgSize,
    nodeWidth,
    largestDagScale,
  ]);

  return {
    svgSize,
    navigateToDag,
  };
};

export default useDag;
