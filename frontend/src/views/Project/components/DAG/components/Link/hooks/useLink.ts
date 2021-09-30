import {line} from 'd3';
import {useState, useCallback, useMemo, useRef} from 'react';

import {
  DagDirection,
  Link,
  PointCoordinates,
  Node,
} from '@dash-frontend/lib/types';
import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import useRouteController from 'hooks/useRouteController';

const HOVER_LINK_THRESHOLD = 250;
const FF_OFFSET_SHORT = 75;
const FF_OFFSET_LONG = 100;

const getLineArray = (
  startPoint: PointCoordinates,
  endPoint: PointCoordinates,
  bendPoints: PointCoordinates[],
  isSelected: boolean,
  dagDirection: DagDirection,
) => {
  const lineArray = bendPoints.reduce<[number, number][]>(
    (acc, point) => {
      acc.push([point.x, point.y]);
      return acc;
    },
    [[startPoint.x, startPoint.y]],
  );

  const xOffset =
    dagDirection === DagDirection.RIGHT ? (isSelected ? 7.5 : 5) : 0;
  const yOffset =
    dagDirection === DagDirection.DOWN ? (isSelected ? 7.5 : 5) : 0;

  lineArray.push([endPoint.x - xOffset, endPoint.y - yOffset]);
  return lineArray;
};

const useLink = (
  link: Link,
  isInteractive: boolean,
  dagDirection: DagDirection,
  nodes: Node[],
) => {
  const pathRef = useRef<SVGPathElement>(null);
  const {selectedNode: selectedNodeId, navigateToNode} = useRouteController();
  const [isHovering, setIsHovering] = useState(false);
  const {hoveredNode} = useHoveredNode();

  const isSelected = useMemo(
    () =>
      (isInteractive && [selectedNodeId, hoveredNode].includes(link.source)) ||
      [selectedNodeId, hoveredNode].includes(link.target),
    [link, selectedNodeId, isInteractive, hoveredNode],
  );
  const d = useMemo(
    () =>
      line()(
        getLineArray(
          link.startPoint,
          link.endPoint,
          link.bendPoints,
          isSelected,
          dagDirection,
        ),
      ) || '',
    [link.startPoint, link.endPoint, link.bendPoints, isSelected, dagDirection],
  );

  const selectedNode = useMemo(
    () => nodes.find((n) => n.id === selectedNodeId),
    [nodes, selectedNodeId],
  );

  const oppositeNode = useMemo(() => {
    if (selectedNodeId === link.source) {
      return nodes.find((n) => n.id === link.target);
    } else if (selectedNodeId === link.target) {
      return nodes.find((n) => n.id === link.source);
    }
  }, [link, nodes, selectedNodeId]);

  const diff = (pathRef.current?.getTotalLength() || 0) > HOVER_LINK_THRESHOLD;
  const sourceSelected = link.target === selectedNodeId;
  const shouldShowClickableLink =
    (link.source === selectedNodeId || sourceSelected) &&
    oppositeNode &&
    diff &&
    isInteractive;

  const onClick = useCallback(() => {
    oppositeNode && navigateToNode(oppositeNode);
  }, [oppositeNode, navigateToNode]);

  const fastForwardIconPoint = useCallback(() => {
    let offset = 0;
    if (dagDirection === DagDirection.RIGHT) {
      offset = !sourceSelected
        ? FF_OFFSET_SHORT
        : (pathRef.current?.getTotalLength() || 0) - FF_OFFSET_LONG;
    }
    if (dagDirection === DagDirection.DOWN) {
      offset = !sourceSelected
        ? FF_OFFSET_LONG
        : (pathRef.current?.getTotalLength() || 0) - FF_OFFSET_SHORT;
    }
    return pathRef.current?.getPointAtLength(offset) || {x: 0, y: 0};
  }, [dagDirection, sourceSelected]);

  const fastForwardIconOrientation = useCallback(
    (iconPoint) => {
      if (dagDirection === DagDirection.RIGHT) {
        return sourceSelected
          ? `rotate(180, ${iconPoint.x}, ${iconPoint.y})`
          : undefined;
      }
      if (dagDirection === DagDirection.DOWN) {
        return `rotate(${sourceSelected ? 270 : 90}, ${iconPoint.x}, ${
          iconPoint.y
        })`;
      }
    },
    [dagDirection, sourceSelected],
  );

  return {
    d,
    isSelected,
    selectedNode,
    oppositeNode,
    transferring: link.transferring,
    shouldShowClickableLink,
    sourceSelected,
    onClick,
    fastForwardIconPoint,
    fastForwardIconOrientation,
    isHovering,
    setIsHovering,
    pathRef,
  };
};

export default useLink;
