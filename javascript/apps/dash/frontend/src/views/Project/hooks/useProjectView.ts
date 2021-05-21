import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {
  D3ZoomEvent,
  zoom as d3Zoom,
  zoomIdentity,
  ZoomTransform,
} from 'd3-zoom';
import flatten from 'lodash/flatten';
import {useEffect, useRef, useState} from 'react';
import {useParams, useHistory} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {DagDirection} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

const SIDEBAR_WIDTH = 384;

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const [svgSize] = useState({
    height: Math.max(300, window.innerHeight - 100),
    width: window.innerWidth,
  });
  const {selectedNode} = useRouteController();
  const {projectId} = useParams<{projectId: string}>();
  const {viewState, setUrlFromViewState} = useUrlQueryState();
  const browserHistory = useHistory();
  const [sliderZoomValue, setSliderZoomValue] = useState(1);
  const zoomRef = useRef<d3.ZoomBehavior<SVGSVGElement, unknown> | null>(null);

  const dagDirection = viewState.dagDirection || DagDirection.RIGHT;

  const rotateDag = () => {
    switch (dagDirection) {
      case DagDirection.DOWN:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.LEFT}),
        );
        break;
      case DagDirection.LEFT:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.UP}),
        );
        break;
      case DagDirection.UP:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.RIGHT}),
        );
        break;
      case DagDirection.RIGHT:
        browserHistory.push(
          setUrlFromViewState({dagDirection: DagDirection.DOWN}),
        );
        break;
    }
  };

  const {dags, loading, error} = useProjectDagsData({
    projectId,
    nodeHeight,
    nodeWidth,
    direction: dagDirection,
  });

  const applyZoom = (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
    const {transform} = event;

    select<SVGGElement, unknown>('#Dags').attr(
      'transform',
      transform.toString(),
    );
    setSliderZoomValue(transform.k);
  };

  // initialize zoom and dags positioning
  useEffect(() => {
    const horizontal =
      dagDirection === DagDirection.RIGHT || dagDirection === DagDirection.LEFT;
    const nodes = flatten((dags || []).map((dag) => dag.nodes));
    const xExtent = extent(nodes, (n) => n.x);
    const yExtent = extent(nodes, (n) => n.y);
    const xMin = xExtent[0] || 0;
    const xMax = xExtent[1] || svgSize.width;
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgSize.height;
    const xScale = svgSize.width / (xMax - xMin);
    const yScale = svgSize.height / 2 / (yMax - yMin);
    const minScale = Math.max(0.6, Math.min(xScale, yScale, 1.5));
    const svg = select<SVGSVGElement, unknown>('#Svg');
    let transform: ZoomTransform;

    if (!zoomRef.current) {
      zoomRef.current = d3Zoom<SVGSVGElement, unknown>()
        .scaleExtent([0.6, 1.5])
        .on('zoom', applyZoom);
    }

    svg.call(zoomRef.current);

    if (selectedNode) {
      const centerNode = nodes.find((n) => n.name === selectedNode);
      const selectedNodeCenterX =
        (svgSize.width - SIDEBAR_WIDTH) / 2 - (centerNode?.x || 0) * 1.5;
      const selectedNodeCenterY =
        svgSize.height / 2 - (centerNode?.y || 0) * 1.5 - nodeHeight;

      transform = zoomIdentity
        .translate(selectedNodeCenterX, selectedNodeCenterY)
        .scale(1.5);
    } else {
      transform = zoomIdentity
        .translate(
          horizontal ? 0 : svgSize.width / 2,
          horizontal ? svgSize.height / 2 : 0,
        )
        .translate(
          horizontal ? 0 : -(xMin * minScale + xMax * minScale) / 2 - nodeWidth,
          horizontal ? -(yMin * minScale + yMax * minScale) / 2 : 0,
        )
        .scale(minScale);
    }

    svg
      .transition()
      .duration(selectedNode ? 250 : 0)
      .call(zoomRef.current.transform, transform);

    setSliderZoomValue(minScale);
  }, [
    setSliderZoomValue,
    dags,
    svgSize,
    dagDirection,
    nodeWidth,
    nodeHeight,
    selectedNode,
  ]);

  const applySliderZoom = (e: React.FormEvent<HTMLInputElement>) => {
    const nextScale = Number(e.currentTarget.value) / 100;
    zoomRef.current &&
      zoomRef.current.scaleTo(
        select<SVGSVGElement, unknown>('#Svg'),
        nextScale,
      );
  };

  return {
    applySliderZoom,
    dagDirection,
    dags,
    error,
    loading,
    rotateDag,
    sliderZoomValue,
    svgSize,
  };
};
