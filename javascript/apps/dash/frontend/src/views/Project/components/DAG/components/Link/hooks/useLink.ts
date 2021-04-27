import {line} from 'd3';
import {useMemo} from 'react';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {JobState, Link, PointCoordinates} from '@graphqlTypes';

import useRouteController from '../../../hooks/useRouteController';

const getLineArray = (
  startPoint: PointCoordinates,
  endPoint: PointCoordinates,
  bendPoints: PointCoordinates[],
) => {
  const lineArray = bendPoints.reduce<[number, number][]>(
    (acc, point) => {
      acc.push([point.x, point.y]);
      return acc;
    },
    [[startPoint.x, startPoint.y]],
  );

  lineArray.push([endPoint.x, endPoint.y]);
  return lineArray;
};

const useLink = (link: Link) => {
  const {selectedNode} = useRouteController();
  const {hoveredNode} = useHoveredNode();
  const d = useMemo(
    () =>
      line()(getLineArray(link.startPoint, link.endPoint, link.bendPoints)) ||
      '',
    [link.startPoint, link.endPoint, link.bendPoints],
  );

  const transferring = useMemo(() => link.state === JobState.JOB_RUNNING, [
    link.state,
  ]);

  return {
    d,
    hoveredNode,
    selectedNode,
    transferring,
  };
};

export default useLink;
