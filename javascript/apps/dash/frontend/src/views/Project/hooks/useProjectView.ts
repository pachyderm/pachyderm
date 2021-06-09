import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom, zoomIdentity} from 'd3-zoom';
import flatten from 'lodash/flatten';
import {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {useParams, useHistory} from 'react-router';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {DagDirection, Node} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

const SIDEBAR_WIDTH = 384;
export const MAX_SCALE_VALUE = 1.5;
const DEFAULT_MINIMUM_SCALE_VALUE = 0.6;

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
  const [minScale, setMinScale] = useState(DEFAULT_MINIMUM_SCALE_VALUE);
  const [interacted, setInteracted] = useState(false);
  const zoomRef = useRef<d3.ZoomBehavior<SVGSVGElement, unknown> | null>(null);

  const dagDirection = viewState.dagDirection || DagDirection.RIGHT;

  const rotateDag = () => {
    // Reset interaction on rotations, in the future we might want to look into
    // adjusting the current translation on roatation.
    setInteracted(false);
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

  const graphExtents = useMemo(() => {
    const nodes = flatten((dags || []).map((dag) => dag.nodes));
    const xExtent = extent(nodes, (n) => n.x);
    const yExtent = extent(nodes, (n) => n.y);
    const xMin = xExtent[0] || 0;
    const xMax = xExtent[1] || svgSize.width;
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgSize.height;

    return {xMin, xMax, yMin, yMax};
  }, [dags, svgSize]);

  const getScale = (
    svgMeasurement: number,
    max: number,
    min: number,
    padding: number,
  ) => {
    const dagSize = max - min + padding;
    return dagSize > 0 ? svgMeasurement / dagSize : 1;
  };

  const startScale = useMemo(() => {
    const {xMax, xMin, yMax, yMin} = graphExtents;
    const horizontal =
      dagDirection === DagDirection.RIGHT || dagDirection === DagDirection.LEFT;

    // multiply node dimensions by 2 to account for alignment padding
    const xScale = getScale(
      svgSize.width,
      xMax,
      xMin,
      nodeWidth * (horizontal ? 2 : 1),
    );
    const yScale = getScale(
      svgSize.height,
      yMax,
      yMin,
      nodeHeight * (horizontal ? 1 : 2),
    );

    // 0.6 is the largest value allowed for the minimum zoom value
    return Math.max(DEFAULT_MINIMUM_SCALE_VALUE, Math.min(xScale, yScale, 1.5));
  }, [dagDirection, graphExtents, nodeHeight, nodeWidth, svgSize]);

  const applyZoom = (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
    const {transform} = event;

    select<SVGGElement, unknown>('#Dags').attr(
      'transform',
      transform.toString(),
    );
    setSliderZoomValue(transform.k);

    // capture interaction for mousewheel and panning events
    if (event.sourceEvent) setInteracted(true);
  };

  const centerDag = useCallback(
    (isZoomOut = false) => {
      if (zoomRef.current) {
        const svg = select<SVGSVGElement, unknown>('#Svg');
        const horizontal =
          dagDirection === DagDirection.RIGHT ||
          dagDirection === DagDirection.LEFT;
        const {xMin, xMax, yMin, yMax} = graphExtents;

        // translate to center of svg and align based on direction
        const transform = zoomIdentity
          .translate(
            horizontal
              ? nodeWidth
              : svgSize.width / 2 - (xMin + xMax) / 2 - nodeWidth,
            horizontal
              ? svgSize.height / 2 - (yMin + yMax) / 2 - nodeHeight
              : nodeHeight,
          )
          .scale(startScale);

        // zoom.transform does not obey the constraints set on panning and zooming,
        // if constraints are added this should be updated to use one of the methods that obeys them.
        svg
          .transition()
          .duration(isZoomOut ? 250 : 0)
          .call(zoomRef.current.transform, transform)
          .end()
          .then(() => setInteracted(false));

        setSliderZoomValue(startScale);
      }
    },
    [dagDirection, graphExtents, nodeHeight, nodeWidth, startScale, svgSize],
  );

  //initialize zoom and set minimum scale as dags update
  useEffect(() => {
    if (!zoomRef.current) {
      zoomRef.current = d3Zoom<SVGSVGElement, unknown>().on('zoom', applyZoom);
    }

    // 0.6 is the largest value allowed for the minimum zoom value
    zoomRef.current.scaleExtent([
      Math.min(startScale, DEFAULT_MINIMUM_SCALE_VALUE),
      MAX_SCALE_VALUE,
    ]);

    select<SVGSVGElement, unknown>('#Svg').call(zoomRef.current);

    setMinScale(Math.min(startScale, DEFAULT_MINIMUM_SCALE_VALUE));
  }, [startScale, svgSize.height, svgSize.width]);

  // center dag or apply last translation if interacted with when dags update
  useEffect(() => {
    const svg = select<SVGSVGElement, unknown>('#Svg');

    if (!selectedNode && zoomRef.current) {
      // center and align dag if the user has not interacted with it yet
      if (!interacted) {
        centerDag();

        // if the user has interacted apply previous transform against new constraints
      } else {
        const dagsNode = select<SVGGElement, unknown>('#Dags').node();
        if (dagsNode) {
          zoomRef.current.scaleBy(svg, 1);

          zoomRef.current.translateBy(svg, 0, 0);
        }
      }
    }
  }, [centerDag, interacted, selectedNode]);

  // zoom and pan to selected node
  useEffect(() => {
    const centerNodeSelection = select<SVGGElement, Node>(
      `#${selectedNode}GROUP`,
    );

    if (!centerNodeSelection.empty() && zoomRef.current && !loading) {
      setInteracted(true);
      const svg = select<SVGSVGElement, unknown>('#Svg');

      const centerNode = centerNodeSelection.data()[0];

      const selectedNodeCenterX =
        (svgSize.width - SIDEBAR_WIDTH) / 2 -
        (centerNode.x + nodeWidth / 2) * MAX_SCALE_VALUE;
      const selectedNodeCenterY =
        svgSize.height / 2 - (centerNode.y + nodeHeight / 2) * MAX_SCALE_VALUE;

      const transform = zoomIdentity
        .translate(selectedNodeCenterX, selectedNodeCenterY)
        .scale(MAX_SCALE_VALUE);

      // zoom.transform does not obey the constraints set on panning and zooming,
      // if constraints are added this should be updated to use one of the methods that obeys them.
      zoomRef.current.transform(svg.transition(), transform);
    }
  }, [loading, nodeHeight, nodeWidth, selectedNode, svgSize]);

  // reset interaction on empty canvas
  useEffect(() => {
    if (dags && dags.length === 0) setInteracted(false);
  }, [dags]);

  const applySliderZoom = (e: React.FormEvent<HTMLInputElement>) => {
    setInteracted(true);
    const nextScale = Number(e.currentTarget.value) / 100;
    zoomRef.current &&
      zoomRef.current.scaleTo(
        select<SVGSVGElement, unknown>('#Svg'),
        nextScale,
      );
  };

  const zoomOut = useCallback(() => {
    centerDag(true);
  }, [centerDag]);

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
