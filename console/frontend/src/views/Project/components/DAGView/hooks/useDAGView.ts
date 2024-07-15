import {extent} from 'd3-array';
import {select} from 'd3-selection';
import {D3ZoomEvent, zoom as d3Zoom, zoomIdentity} from 'd3-zoom';
import objectHash from 'object-hash';
import {
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  useState,
} from 'react';

import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import {DEFAULT_SIDEBAR_SIZE} from '@dash-frontend/hooks/useSidebarInfo';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {DagDirection, Node, Dags} from '@dash-frontend/lib/types';

import {NODE_HEIGHT, NODE_WIDTH} from '../../../constants/nodeSizes';

import useDAGRouteController from './useDAGRouteController';

export const MAX_SCALE_VALUE = 1.5;
const CENTER_SCALE_VALUE = 1;
const DEFAULT_MINIMUM_SCALE_VALUE = 0.6;
const DAG_CONTROLS_HEIGHT = 65;

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

export const useDAGView = (
  nodeWidth: number,
  nodeHeight: number,
  dags: Dags | undefined,
  loading: boolean,
) => {
  const dagRef = useRef<SVGSVGElement>(null);
  const {selectedNode} = useDAGRouteController();
  const {searchParams} = useUrlQueryState();
  const {pipelineId, repoId, projectId} = useUrlState();
  const {currentProject} = useCurrentProject();
  const [dagDirectionSetting, setDagDirectionSetting] = useLocalProjectSettings(
    {projectId, key: 'dag_direction'},
  );
  const [skipCenterOnSelect, setSkipCenterOnSelectSetting] =
    useLocalProjectSettings({projectId, key: 'skip_center_on_select'});
  const [sidebarWidthSetting] = useLocalProjectSettings({
    projectId,
    key: 'sidebar_width',
  });
  const sidebarSize = Number(sidebarWidthSetting || DEFAULT_SIDEBAR_SIZE);
  const [dagState, dispatch] = useReducer(dagReducer, {
    interacted: false,
    reset: false,
  });
  const [sliderZoomValue, setSliderZoomValue] = useState(1);
  const [minScale, setMinScale] = useState(DEFAULT_MINIMUM_SCALE_VALUE);
  const zoomRef = useRef<d3.ZoomBehavior<SVGSVGElement, unknown> | null>(null);

  const dagDirection = dagDirectionSetting || DagDirection.DOWN;

  const {interacted, reset} = dagState;

  const handleChangeCenterOnSelect = (shouldCenter: boolean) => {
    setSkipCenterOnSelectSetting(!shouldCenter);
  };

  const rotateDag = useCallback(() => {
    // Reset interaction on rotations, in the future we might want to look into
    // adjusting the current translation on rotation.
    dispatch({type: 'ROTATE'});

    const handleChangeDirection = (nextDirection: DagDirection) => {
      setDagDirectionSetting(nextDirection);
    };

    switch (dagDirection) {
      case DagDirection.DOWN:
        handleChangeDirection(DagDirection.RIGHT);
        break;
      case DagDirection.RIGHT:
        handleChangeDirection(DagDirection.DOWN);
        break;
    }
  }, [dagDirection, setDagDirectionSetting]);

  const graphExtents = useMemo(() => {
    const nodes = dags?.nodes || [];
    const xExtent = extent(nodes, (n) => n.x);
    const yExtent = extent(nodes, (n) => n.y);
    const xMin = xExtent[0] || 0;
    const xMax = (xExtent[1] || 0) + NODE_WIDTH;
    const yMin = yExtent[0] || 0;
    const yMax = (yExtent[1] || 0) + NODE_HEIGHT;

    return {xMin, xMax, yMin, yMax};
  }, [dags]);

  const getScale = (svgMeasurement: number, max: number, padding: number) => {
    const dagSize = max + padding;
    return dagSize > 0 ? svgMeasurement / dagSize : 1;
  };

  // get the scale needed to display the entire dag within the svg container
  const getMinimumScale = useCallback(() => {
    const {xMax, yMax} = graphExtents;
    if (dagRef.current) {
      const svgSize = dagRef.current?.getBoundingClientRect() || {
        width: 0,
        height: 0,
      };

      const xScale = getScale(svgSize?.width, xMax, nodeWidth);
      const yScale = getScale(
        svgSize?.height - DAG_CONTROLS_HEIGHT,
        yMax,
        nodeHeight,
      );
      return Math.min(xScale, yScale, MAX_SCALE_VALUE);
    } else return 1;
  }, [graphExtents, nodeHeight, nodeWidth, dagRef]);

  const applyZoom = useCallback(
    (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
      const {transform} = event;

      if (!isNaN(transform.k)) {
        select<SVGGElement, unknown>('#Dags').attr(
          'transform',
          transform.toString(),
        );
        setSliderZoomValue(transform.k);
      }

      // capture interaction for mousewheel and panning events
      if (event.sourceEvent) dispatch({type: 'PAN'});
    },
    [],
  );

  const centerDag = useCallback(() => {
    if (zoomRef.current && dagRef.current) {
      const svg = select<SVGSVGElement, unknown>('#Svg');
      const horizontal = dagDirection === DagDirection.RIGHT;
      const {xMin, xMax, yMin, yMax} = graphExtents;
      const svgSize = dagRef.current?.getBoundingClientRect();
      const startScale = getMinimumScale();

      const verticalDAGYTranslate =
        nodeHeight * startScale > DAG_CONTROLS_HEIGHT
          ? nodeHeight * startScale
          : DAG_CONTROLS_HEIGHT;

      const yTranslate = horizontal
        ? (svgSize.height + DAG_CONTROLS_HEIGHT) / 2 -
          ((yMin + yMax) / 2) * startScale
        : verticalDAGYTranslate;

      const xTranslate = horizontal
        ? (nodeWidth * startScale) / 2
        : svgSize.width / 2 - ((xMin + xMax) / 2) * startScale;

      // translate to center of svg and align based on direction
      const transform = zoomIdentity
        .translate(xTranslate, yTranslate)
        .scale(startScale);

      // zoom.transform does not obey the constraints set on panning and zooming,
      // if constraints are added this should be updated to use one of the methods that obeys them.
      svg.transition().duration(250).call(zoomRef.current.transform, transform);
      setSliderZoomValue(startScale);
    }
  }, [dagDirection, getMinimumScale, graphExtents, nodeHeight, nodeWidth]);

  //initialize zoom and set minimum scale as dags update
  useEffect(() => {
    const startScale = getMinimumScale();
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
  }, [applyZoom, loading, getMinimumScale, centerDag]);

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
    if (!skipCenterOnSelect) {
      let x, y;
      const hash = objectHash({
        project: currentProject?.project?.name,
        name: selectedNode,
      });

      const nodeInView = select<SVGGElement, Node>(`#GROUP_${hash}`);
      if (!nodeInView.empty()) {
        x = nodeInView.data()[0].x;
        y = nodeInView.data()[0].y;
      } else {
        const offscreenNode = dags?.nodes.find(({id}) => id === hash);
        x = offscreenNode?.x;
        y = offscreenNode?.y;
      }

      if (x && y && zoomRef.current && !loading && interacted && !reset) {
        const svg = select<SVGSVGElement, unknown>('#Svg');
        const svgSize = dagRef.current
          ? dagRef.current?.getBoundingClientRect()
          : {
              width: 0,
              height: 0,
            };

        const selectedNodeCenterX =
          (svgSize.width - sidebarSize) / 2 -
          (x + nodeWidth / 2) * CENTER_SCALE_VALUE;
        const selectedNodeCenterY =
          svgSize.height / 2 - (y + nodeHeight / 2) * CENTER_SCALE_VALUE;

        const transform = zoomIdentity
          .translate(selectedNodeCenterX, selectedNodeCenterY)
          .scale(CENTER_SCALE_VALUE);

        const transition = svg.transition().on('interrupt', () => {
          zoomRef.current?.transform(svg.transition(), transform);
        });
        // zoom.transform does not obey the constraints set on panning and zooming,
        // if constraints are added this should be updated to use one of the methods that obeys them.
        zoomRef.current.transform(transition, transform);
      }
    }
  }, [
    loading,
    nodeHeight,
    nodeWidth,
    selectedNode,
    interacted,
    reset,
    skipCenterOnSelect,
    currentProject?.project?.name,
    dags?.nodes,
    sidebarWidthSetting,
    sidebarSize,
  ]);

  // reset interaction on empty canvas
  useEffect(() => {
    if (dags && dags.links.length === 0 && dags.nodes.length === 0) {
      dispatch({type: 'EMPTY'});
    }
  }, [dags]);

  const applySliderZoom = useCallback((value: number) => {
    dispatch({type: 'ZOOM'});
    const nextScale = value / 100;
    zoomRef.current &&
      zoomRef.current.scaleTo(
        select<SVGSVGElement, unknown>('#Svg'),
        nextScale,
      );
  }, []);

  const zoomOut = useCallback(() => {
    dispatch({type: 'RESET'});
  }, []);

  // DAG controls keyboard interactions
  useEffect(() => {
    const ZOOM_INCREMENT = 10;

    document.onkeydown = (event) => {
      if (event.key === '@') {
        zoomOut();
      }
      if (event.metaKey && event.key === '-') {
        event.preventDefault();
        event.stopPropagation();
        applySliderZoom(sliderZoomValue * 100 - ZOOM_INCREMENT);
      }
      if (event.metaKey && event.key === '=') {
        event.preventDefault();
        event.stopPropagation();
        applySliderZoom(sliderZoomValue * 100 + ZOOM_INCREMENT);
      }
    };

    return () => {
      document.onkeydown = null;
    };
  }, [applySliderZoom, sliderZoomValue, zoomOut]);

  // Always zoomOut when the global id filter changes.
  useEffect(() => {
    if (!loading) {
      dispatch({type: 'RESET'});
    }
  }, [loading, searchParams.globalIdFilter]);

  useEffect(() => {
    if (!loading && (pipelineId || repoId)) {
      dispatch({type: 'SELECT_NODE'});
    }
  }, [pipelineId, repoId, loading]);

  return {
    applySliderZoom,
    minScale,
    rotateDag,
    dagDirection,
    sliderZoomValue,
    zoomOut,
    skipCenterOnSelect,
    handleChangeCenterOnSelect,
    graphExtents,
    projectName: currentProject?.project?.name,
    searchParams,
    dagRef,
  };
};
