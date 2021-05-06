import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom, zoomIdentity} from 'd3-zoom';
import {useEffect, useState, useRef} from 'react';

import {Dag, DagDirection} from '@graphqlTypes';

import useRouteController from './useRouteController';

type useDagProps = {
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  setLargestDagScale: React.Dispatch<React.SetStateAction<number | null>>;
  setSliderZoomValue: React.Dispatch<React.SetStateAction<number>>;
  largestDagScale: number | null;
  dagDirection: DagDirection;
};

const useDag = ({
  id,
  nodeHeight,
  nodeWidth,
  largestDagScale,
  setLargestDagScale,
  data,
  isInteractive,
  setSliderZoomValue,
  dagDirection,
}: useDagProps) => {
  const {navigateToDag, selectedNode} = useRouteController();
  const [svgSize, setSVGSize] = useState({
    width: 0,
    height: 0,
  });

  const zoomRef = useRef<d3.ZoomBehavior<SVGSVGElement, unknown> | null>(null);
  const graphRef = useRef(select<SVGGElement, unknown>(`#${id}Graph`));
  const svgRef = useRef(select<SVGSVGElement, unknown>(`#${id}`));

  const applySliderZoom = (e: React.FormEvent<HTMLInputElement>) => {
    const nextScale = Number(e.currentTarget.value) / 100;
    if (zoomRef.current) {
      zoomRef.current.scaleTo(svgRef.current, nextScale);
    }
  };

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
      setSVGSize({
        height: svgElement?.getBBox().height || 0,
        width: parentWidth,
      });
      const scale = Math.min(1, parentWidth / dagWidth);

      if (!largestDagScale || scale < largestDagScale) {
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
    svgRef.current = select<SVGSVGElement, unknown>(`#${id}`);
    graphRef.current = select<SVGGElement, unknown>(`#${id}Graph`);

    const applyZoom = (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
      const {transform, sourceEvent} = event;
      const node = graphRef.current.node();

      if (node) {
        // grab actual transform values from the graph 'g' element
        const svgPositions = node
          .getAttribute('transform')
          ?.split(/[a-zA-Z() ,]+/)
          .slice(1, 4);
        const [x, y, scale] = svgPositions?.map((n) => Number(n)) || [0, 0, 1];

        const applyScaleOnly = zoomIdentity.translate(x, y).scale(transform.k);

        const applyTranslateOnly = zoomIdentity
          .translate(
            x + sourceEvent?.movementX || 0,
            y + sourceEvent?.movementY || 0,
          )
          .scale(transform.k);

        if (transform.k !== Number(scale) || sourceEvent?.type === 'wheel') {
          // covers wheel scroll and slider scroll events
          graphRef.current.attr('transform', applyScaleOnly.toString());
        } else if (sourceEvent?.type === 'mousemove') {
          // covers mouse drag panning
          graphRef.current.attr('transform', applyTranslateOnly.toString());
        } else {
          // covers initial centering position
          graphRef.current.attr('transform', transform.toString());
        }
        setSliderZoomValue(transform.k);
      }
    };

    const zoom = d3Zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.6, 1])
      .on('zoom', applyZoom);

    svgRef.current.call(zoom);
    zoomRef.current = zoom;

    const yExtent = extent(data.nodes, (d) => d.y);
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgSize.height;
    const scale = (!isInteractive && largestDagScale) || 1;

    const transform = zoomIdentity
      // center dag within parent (32 is the top padding of ${id}Base)
      .translate((nodeWidth / 2) * scale, svgSize.height / 2 - 32)
      .translate(0, -(yMin + yMax) / 2)
      .scale(scale);

    svgRef.current.call(zoom.transform, transform);

    !isInteractive && zoom.on('zoom', null);
  }, [
    data.nodes,
    id,
    isInteractive,
    nodeHeight,
    svgSize,
    nodeWidth,
    largestDagScale,
    setSliderZoomValue,
    selectedNode,
    dagDirection,
  ]);

  return {
    svgSize,
    navigateToDag,
    applySliderZoom,
  };
};

export default useDag;
