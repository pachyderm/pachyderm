import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom, zoomIdentity} from 'd3-zoom';
import flatten from 'lodash/flatten';
import {
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  useState,
} from 'react';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection, Node} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

const SIDEBAR_WIDTH = 384;
const MIN_DAG_HEIGHT = 300;
const DAG_TOP_PADDING = 100;
export const MAX_SCALE_VALUE = 1.5;
const DEFAULT_MINIMUM_SCALE_VALUE = 0.6;

interface DagState {
  interacted: boolean;
  reset: boolean;
}

type DagAction =
  | {type: 'ROTATE'}
  | {type: 'ZOOM'}
  | {type: 'EMPTY'}
  | {type: 'RESET'}
  | {type: 'PAN'}
  | {type: 'SELECT_NODE'};

const dagReducer = (state: DagState, action: DagAction): DagState => {
  switch (action.type) {
    case 'ROTATE':
    case 'RESET':
    case 'EMPTY':
      return {
        ...state,
        interacted: false,
        reset: true,
      };
    case 'ZOOM':
    case 'PAN':
      return {
        ...state,
        interacted: true,
      };
    case 'SELECT_NODE':
      return {
        ...state,
        interacted: true,
        reset: false,
      };
    default:
      return state;
  }
};

export const useProjectView = (nodeWidth: number, nodeHeight: number) => {
  const [svgSize, setSvgSize] = useState({
    height: Math.max(MIN_DAG_HEIGHT, window.innerHeight - DAG_TOP_PADDING),
    width: window.innerWidth,
  });
  const {isOpen, overlay, sidebarSize} = useSidebarInfo();
  const {selectedNode} = useRouteController();
  const {viewState, setUrlFromViewState} = useUrlQueryState();
  const {pipelineId, repoId, projectId, jobId} = useUrlState();
  const [dagState, dispatch] = useReducer(dagReducer, {
    interacted: false,
    reset: false,
  });
  const [sliderZoomValue, setSliderZoomValue] = useState(1);
  const [minScale, setMinScale] = useState(DEFAULT_MINIMUM_SCALE_VALUE);
  const zoomRef = useRef<d3.ZoomBehavior<SVGSVGElement, unknown> | null>(null);

  const dagDirection = viewState.dagDirection || DagDirection.RIGHT;
  const {interacted, reset} = dagState;

  const rotateDag = useCallback(() => {
    // Reset interaction on rotations, in the future we might want to look into
    // adjusting the current translation on rotation.
    dispatch({type: 'ROTATE'});

    switch (dagDirection) {
      case DagDirection.DOWN:
        setUrlFromViewState({
          dagDirection: DagDirection.RIGHT,
        });
        break;
      case DagDirection.RIGHT:
        setUrlFromViewState({
          dagDirection: DagDirection.DOWN,
        });
        break;
    }
  }, [setUrlFromViewState, dagDirection]);

  const {dags, loading, error} = useProjectDagsData({
    jobSetId: jobId,
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
    const dagSize = max + min + padding;
    return dagSize > 0 ? svgMeasurement / dagSize : 1;
  };

  const startScale = useMemo(() => {
    const {xMax, xMin, yMax, yMin} = graphExtents;

    // multiply node dimensions by 2 to account for alignment padding
    const xScale = getScale(svgSize.width, xMax, xMin, nodeWidth);
    const yScale = getScale(svgSize.height, yMax, yMin, nodeHeight);

    return Math.min(xScale, yScale, 1.5);
  }, [graphExtents, nodeHeight, nodeWidth, svgSize]);

  const applyZoom = useCallback(
    (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
      const {transform} = event;

      select<SVGGElement, unknown>('#Dags').attr(
        'transform',
        transform.toString(),
      );
      setSliderZoomValue(transform.k);

      // capture interaction for mousewheel and panning events
      if (event.sourceEvent) dispatch({type: 'PAN'});
    },
    [],
  );

  const centerDag = useCallback(() => {
    if (zoomRef.current) {
      const svg = select<SVGSVGElement, unknown>('#Svg');
      const horizontal = dagDirection === DagDirection.RIGHT;
      const {xMin, xMax, yMin, yMax} = graphExtents;

      const yTranslate = horizontal
        ? svgSize.height / 2 - ((yMin + yMax) / 2) * startScale - nodeHeight
        : nodeHeight * startScale;

      const xTranslate = horizontal
        ? nodeWidth * startScale
        : svgSize.width / 2 - ((xMin + xMax) / 2) * startScale - nodeWidth;
      // translate to center of svg and align based on direction
      const transform = zoomIdentity
        .translate(
          xTranslate < 0 ? xTranslate : xTranslate,
          yTranslate < 0 ? yTranslate : yTranslate,
        )
        .scale(startScale);

      // zoom.transform does not obey the constraints set on panning and zooming,
      // if constraints are added this should be updated to use one of the methods that obeys them.
      svg.transition().duration(250).call(zoomRef.current.transform, transform);

      setSliderZoomValue(startScale);
    }
  }, [dagDirection, graphExtents, nodeHeight, nodeWidth, startScale, svgSize]);

  // set window resize listener for svg parent size
  useEffect(() => {
    const resizeSvg = () =>
      setSvgSize({
        height: Math.max(MIN_DAG_HEIGHT, window.innerHeight - DAG_TOP_PADDING),
        width: window.innerWidth,
      });
    window.addEventListener('resize', resizeSvg, true);
    return () => window.removeEventListener('resize', resizeSvg);
  }, []);

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
    // need to re-apply this effect when loading changes, so that the
    // zoomRef can be updated to reflect the new DAG scale
  }, [startScale, svgSize.height, svgSize.width, applyZoom, loading]);

  // center dag or apply last translation if interacted with when dags update
  useEffect(() => {
    const svg = select<SVGSVGElement, unknown>('#Svg');

    if (zoomRef.current) {
      // center and align dag if the user has not interacted with it yet
      if ((reset && !interacted) || (!reset && !interacted)) {
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
  }, [centerDag, selectedNode, interacted, reset, loading]);

  // zoom and pan to selected node
  useEffect(() => {
    const centerNodeSelection = select<SVGGElement, Node>(
      `#${selectedNode}GROUP`,
    );

    if (
      !centerNodeSelection.empty() &&
      zoomRef.current &&
      !loading &&
      interacted &&
      !reset
    ) {
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
  }, [
    loading,
    nodeHeight,
    nodeWidth,
    selectedNode,
    svgSize,
    interacted,
    reset,
  ]);

  // reset interaction on empty canvas
  useEffect(() => {
    if (dags && dags.length === 0) {
      dispatch({type: 'EMPTY'});
    }
  }, [dags]);

  const applySliderZoom = useCallback(
    (e: React.FormEvent<HTMLInputElement>) => {
      dispatch({type: 'ZOOM'});
      const nextScale = Number(e.currentTarget.value) / 100;
      zoomRef.current &&
        zoomRef.current.scaleTo(
          select<SVGSVGElement, unknown>('#Svg'),
          nextScale,
        );
    },
    [],
  );

  const zoomOut = useCallback(() => {
    dispatch({type: 'RESET'});
  }, []);

  useEffect(() => {
    const zoomOutListener = (event: KeyboardEvent) => {
      if (event.key === '@') {
        zoomOut();
      }
    };
    window.addEventListener('keydown', zoomOutListener);

    return () => window.removeEventListener('keydown', zoomOutListener);
  }, [zoomOut]);

  useEffect(() => {
    if (overlay) {
      dispatch({type: 'RESET'});
    }
  }, [overlay]);

  useEffect(() => {
    if (!loading) {
      dispatch({type: 'RESET'});
    }
  }, [loading]);

  useEffect(() => {
    if (!loading && (pipelineId || repoId)) {
      dispatch({type: 'SELECT_NODE'});
    }
  }, [pipelineId, repoId, loading]);

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
    isSidebarOpen: isOpen,
    sidebarSize,
  };
};
