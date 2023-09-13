import {line} from 'd3';
import {useMemo, useRef} from 'react';

import {DagDirection, Link, PointCoordinates} from '@dash-frontend/lib/types';

const getLineArray = (
  startPoint: PointCoordinates,
  endPoint: PointCoordinates,
  bendPoints: PointCoordinates[],
  dagDirection: DagDirection,
) => {
  const lineArray = bendPoints.reduce<[number, number][]>(
    (acc, point) => {
      acc.push([point.x, point.y]);
      return acc;
    },
    [[startPoint.x, startPoint.y]],
  );

  const xOffset = dagDirection === DagDirection.RIGHT ? 5 : 0;
  const yOffset = dagDirection === DagDirection.DOWN ? 5 : 0;

  lineArray.push([endPoint.x - xOffset, endPoint.y - yOffset]);
  return lineArray;
};

const useLink = (link: Link, dagDirection: DagDirection) => {
  const pathRef = useRef<SVGPathElement>(null);
  const d = useMemo(
    () =>
      line()(
        getLineArray(
          link.startPoint,
          link.endPoint,
          link.bendPoints,
          dagDirection,
        ),
      ) || '',
    [link.startPoint, link.endPoint, link.bendPoints, dagDirection],
  );

  return {
    d,
    transferring: link.transferring,
    isCrossProject: link.isCrossProject,
    pathRef,
  };
};

export default useLink;
