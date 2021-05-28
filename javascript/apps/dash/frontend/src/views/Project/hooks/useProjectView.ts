import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom} from 'd3-zoom';
import flatten from 'lodash/flatten';
import {useCallback, useEffect, useRef, useState} from 'react';
import {useParams, useHistory} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {DagDirection, Node} from '@graphqlTypes';
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
  const [minScale, setMinScale] = useState(0.6);
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
    const startScale = Math.max(0.6, Math.min(xScale, yScale, 1.5));
    const svg = select<SVGSVGElement, unknown>('#Svg');
    const dagsNode = select<SVGGElement, unknown>('#Dags').node();
    if (dagsNode) {
      const {x, y, width, height} = dagsNode.getBBox();
      zoomRef.current = d3Zoom<SVGSVGElement, unknown>()
        .scaleExtent([Math.min(startScale, 0.6), 1.5])
        .extent([
          [0, 0],
          [svgSize.width, svgSize.height],
        ])
        .translateExtent([
          [x, y],
          [x + width, y + height],
        ])
        .constrain((transform, extent, translateExtent) => {
          const dx0 =
              transform.invertX(extent[0][0]) -
              (translateExtent[0][0] - width / 2),
            dx1 =
              transform.invertX(extent[1][0]) -
              (translateExtent[1][0] + width / 2),
            dy0 =
              transform.invertY(extent[0][1]) -
              (translateExtent[0][1] - height / 2),
            dy1 =
              transform.invertY(extent[1][1]) -
              (translateExtent[1][1] + height / 2);

          return transform.translate(
            dx1 > dx0 ? (dx0 + dx1) / 2 : Math.min(0, dx0) || Math.max(0, dx1),
            dy1 > dy0 ? (dy0 + dy1) / 2 : Math.min(0, dy0) || Math.max(0, dy1),
          );
        })
        .on('zoom', applyZoom);

      svg.call(zoomRef.current);

      if (selectedNode) {
        const centerNodeSelection = select<SVGGElement, Node>(
          `#${selectedNode}GROUP`,
        );
        if (!centerNodeSelection.empty()) {
          const centerNode = select<SVGGElement, Node>(
            `#${selectedNode}GROUP`,
          ).data()[0];
          const selectedNodeCenterX =
            (svgSize.width - SIDEBAR_WIDTH) / 2 - centerNode.x * 1.5;
          const selectedNodeCenterY =
            svgSize.height / 2 - centerNode.y * 1.5 - nodeHeight;

          zoomRef.current.scaleTo(svg, 1.5, [
            selectedNodeCenterX,
            selectedNodeCenterY,
          ]);

          zoomRef.current.translateTo(
            svg.transition().duration(250),
            centerNode.x,
            centerNode.y,
            [(svgSize.width - SIDEBAR_WIDTH) / 2, svgSize.height / 2],
          );
        }
      } else {
        zoomRef.current.scaleTo(svg, startScale);
        zoomRef.current.translateTo(svg, 0, 0, [
          horizontal ? nodeWidth : svgSize.width / 2,
          horizontal ? svgSize.height / 2 : nodeHeight,
        ]);
        zoomRef.current.translateBy(
          svg,
          horizontal ? 0 : -(xMin + xMax) / 2 - nodeWidth,
          horizontal ? -(yMin + yMax) / 2 : 0,
        );

        setSliderZoomValue(startScale);
        setMinScale(Math.min(startScale, 0.6));
      }
    }
  }, [
    setMinScale,
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

  const zoomOut = useCallback(() => {
    zoomRef.current &&
      select<SVGSVGElement, unknown>('#Svg')
        .transition()
        .duration(250)
        .call(zoomRef.current.scaleTo, minScale);
  }, [minScale]);

  useEffect(() => {
    const zoomOutListener = (event: KeyboardEvent) => {
      if (event.key === '@') {
        zoomOut();
      }
    };
    window.addEventListener('keydown', zoomOutListener);

    return () => window.removeEventListener('keydown', zoomOutListener);
  }, [zoomOut]);

  return {
    applySliderZoom,
    dagDirection,
    dags,
    error,
    loading,
    minScale,
    rotateDag,
    sliderZoomValue,
    svgSize,
    zoomOut,
  };
};
