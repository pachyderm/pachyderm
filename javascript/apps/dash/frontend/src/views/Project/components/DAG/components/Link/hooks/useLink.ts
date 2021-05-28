import {line} from 'd3';
import {useMemo} from 'react';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {Link, PointCoordinates} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';

const getLineArray = (
  startPoint: PointCoordinates,
  endPoint: PointCoordinates,
  bendPoints: PointCoordinates[],
  offset: {x: number; y: number},
) => {
  const lineArray = bendPoints.reduce<[number, number][]>(
    (acc, point) => {
      acc.push([point.x + offset.x, point.y + offset.y]);
      return acc;
    },
    [[startPoint.x + offset.x, startPoint.y + offset.y]],
  );

  lineArray.push([endPoint.x + offset.x, endPoint.y + offset.y]);
  return lineArray;
};

const useLink = (link: Link, offset: {x: number; y: number}) => {
  const {selectedNode} = useRouteController();
  const {hoveredNode} = useHoveredNode();
  const d = useMemo(
    () =>
      line()(
        getLineArray(link.startPoint, link.endPoint, link.bendPoints, offset),
      ) || '',
    [link.startPoint, link.endPoint, link.bendPoints, offset],
  );

  return {
    d,
    hoveredNode,
    selectedNode,
    transferring: link.transferring,
  };
};

export default useLink;
