import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom, zoomIdentity} from 'd3-zoom';
import {useEffect, useState, useRef} from 'react';

import {Dag, DagDirection} from '@graphqlTypes';

import useRouteController from './useRouteController';

const SIDEBAR_WIDTH = 384;
const CANVAS_CONTROLS_HEIGHT = 64;
const THUMBNAIL_TOP_PADDING = 32;

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

  // Set refs, initialize zoom and zoom handler
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
        const [x, y] = svgPositions?.map((n) => Number(n)) || [0, 0, 1];

        const applyScaleOnly = zoomIdentity.translate(x, y).scale(transform.k);

        const applyTransform = zoomIdentity
          .translate(
            x + sourceEvent?.movementX || x,
            y + sourceEvent?.movementY || y,
          )
          .scale(transform.k);

        if (sourceEvent?.type === 'wheel') {
          graphRef.current.attr('transform', applyScaleOnly.toString());
        } else {
          graphRef.current.attr('transform', applyTransform.toString());
        }
        setSliderZoomValue(transform.k);
      }
    };

    const zoom = d3Zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.6, 1])
      .on('zoom', applyZoom);

    svgRef.current.call(zoom);
    zoomRef.current = zoom;
  }, [data.nodes, id, isInteractive, setSliderZoomValue]);

  // Left align and vertically center dag, or center on selected node
  useEffect(() => {
    const yExtent = extent(data.nodes, (d) => d.y);
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgSize.height;
    const scale = (!isInteractive && largestDagScale) || 1;

    let transform = zoomIdentity
      .translate(
        (nodeWidth / 2) * scale,
        svgSize.height / 2 - THUMBNAIL_TOP_PADDING,
      )
      .translate(0, -(yMin + yMax) / 2)
      .scale(scale);

    if (selectedNode) {
      const centerNode = data.nodes.find((n) => n.name === selectedNode);

      const selectedNodeCenterX =
        (svgSize.width - SIDEBAR_WIDTH) / 2 - (centerNode?.x || 0);
      const selectedNodeCenterY =
        (svgSize.height - CANVAS_CONTROLS_HEIGHT) / 2 -
        (centerNode?.y || 0) -
        nodeHeight;

      transform = zoomIdentity
        .translate(selectedNodeCenterX, selectedNodeCenterY)
        .scale(scale);

      // the zoom handler is set to ignore translations on programmatic events since they're inaccurate,
      // so we use attr directly to update the translation on node selection.
      graphRef.current
        .transition()
        .duration(250)
        .attr('transform', transform.toString());

      zoomRef.current &&
        svgRef.current
          .transition()
          .duration(250)
          .call(zoomRef.current.transform, transform);
    } else {
      if (zoomRef.current) {
        graphRef.current.attr('transform', transform.toString());
        svgRef.current.call(zoomRef.current.transform, transform);

        if (!isInteractive) {
          zoomRef.current.on('zoom', null);
        }
      }
    }
    return () => {
      zoomRef.current &&
        svgRef.current.call(zoomRef.current.transform, transform);
    };
  }, [
    data.nodes,
    isInteractive,
    nodeHeight,
    svgSize,
    nodeWidth,
    largestDagScale,
    selectedNode,
  ]);

  return {
    svgSize,
    navigateToDag,
    applySliderZoom,
  };
};

export default useDag;
