import {line} from 'd3';
import {useMemo} from 'react';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {DagDirection, Link, PointCoordinates} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

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
) => {
  const {selectedNode} = useRouteController();
  const {hoveredNode} = useHoveredNode();

  const isSelected = useMemo(
    () =>
      (isInteractive && [selectedNode, hoveredNode].includes(link.source)) ||
      [selectedNode, hoveredNode].includes(link.target),
    [link, selectedNode, isInteractive, hoveredNode],
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

  return {
    d,
    hoveredNode,
    isSelected,
    selectedNode,
    transferring: link.transferring,
  };
};

export default useLink;
