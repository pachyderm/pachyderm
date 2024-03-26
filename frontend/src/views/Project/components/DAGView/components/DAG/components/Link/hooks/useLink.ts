import {line, select} from 'd3';
import {useCallback, useMemo} from 'react';

import {DagDirection, Link, PointCoordinates} from '@dash-frontend/lib/types';
import useHoveredNode from '@dash-frontend/views/Project/components/DAGView/providers/HoveredNodeProvider/hooks/useHoveredNode';

const DESIRED_SPEED = 60;
const MASK_WIDTH = 2;
//The marker is currently defined in DAGView.tsx
const MARKER_WIDTH = 4;
const MARKER_LENGTH = 10;

export const CIRCLE_RADIUS = 10;

export const getLineArray = (
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

  const xOffset = dagDirection === DagDirection.RIGHT ? MARKER_LENGTH : 0;
  const yOffset = dagDirection === DagDirection.DOWN ? MARKER_LENGTH : 0;

  lineArray.push([endPoint.x - xOffset, endPoint.y - yOffset]);
  return lineArray;
};

// This function is used to create a polygon around the path that elk generates.
// using a clip path is more performant than using a mask for SVGs. This code takes
// the path and creates a buffer around it of a given width. We need to detect what
// type of turn the line does at a given point and correctly add the offset to the buffer.
export const getLineClipPath = (
  startPoint: PointCoordinates,
  endPoint: PointCoordinates,
  bendPoints: PointCoordinates[],
  dagDirection: DagDirection,
) => {
  const bufferSize = MASK_WIDTH / 2;

  const allPoints = [startPoint, ...bendPoints, endPoint];

  // We need to create a left and right buffer around the line segment.
  // Based on the direction of the angle we will add or subtract from the
  // current x and y coordinates to create a buffer.
  const leftPath: [number, number][] = [];
  const rightPath: [number, number][] = [];

  allPoints.forEach((point, index, array) => {
    if (index !== 0 && index !== array.length - 1) {
      // a, b and c are three points that we use to determine the direction
      // of the angel. Point b is the point that we are solving for
      // a .
      //   |
      //   |
      // b . ----> . c
      const a = array[index - 1];
      const b = point;
      const c = array[index + 1];

      if (a.x < b.x && a.y === b.y) {
        // Right
        if (b.y < c.y) {
          // Down
          leftPath.push([point.x - bufferSize, point.y + bufferSize]);
          rightPath.push([point.x + bufferSize, point.y - bufferSize]);
        } else {
          // Up
          leftPath.push([point.x + bufferSize, point.y + bufferSize]);
          rightPath.push([point.x - bufferSize, point.y - bufferSize]);
        }
      } else if (a.x > b.x && a.y === b.y) {
        // Left
        if (b.y < c.y) {
          // Down
          leftPath.push([point.x - bufferSize, point.y - bufferSize]);
          rightPath.push([point.x + bufferSize, point.y + bufferSize]);
        } else {
          // Up
          leftPath.push([point.x + bufferSize, point.y - bufferSize]);
          rightPath.push([point.x - bufferSize, point.y + bufferSize]);
        }
      } else if (a.y < b.y && a.x === b.x) {
        // Down
        if (b.x < c.x) {
          // Right
          leftPath.push([point.x - bufferSize, point.y + bufferSize]);
          rightPath.push([point.x + bufferSize, point.y - bufferSize]);
        } else {
          // Left
          leftPath.push([point.x - bufferSize, point.y - bufferSize]);
          rightPath.push([point.x + bufferSize, point.y + bufferSize]);
        }
      } else if (a.y > b.y && a.x === b.x) {
        // Up
        if (b.x < c.x) {
          // Right
          leftPath.push([point.x + bufferSize, point.y + bufferSize]);
          rightPath.push([point.x - bufferSize, point.y - bufferSize]);
        } else {
          // Left
          leftPath.push([point.x + bufferSize, point.y - bufferSize]);
          rightPath.push([point.x - bufferSize, point.y + bufferSize]);
        }
      }
    }
  });

  // End marker path. We need to add the arrow shape to the end of the clip path
  const xOffset = dagDirection === DagDirection.RIGHT ? MARKER_LENGTH : 0;
  const yOffset = dagDirection === DagDirection.DOWN ? MARKER_LENGTH : 0;
  const adjustedEndPoint = {x: endPoint.x - xOffset, y: endPoint.y - yOffset};

  const endMarker: [number, number][] = [
    dagDirection === DagDirection.DOWN
      ? [adjustedEndPoint.x - bufferSize, adjustedEndPoint.y]
      : [adjustedEndPoint.x, adjustedEndPoint.y + bufferSize],
    dagDirection === DagDirection.DOWN
      ? [adjustedEndPoint.x - bufferSize - MARKER_WIDTH, adjustedEndPoint.y]
      : [adjustedEndPoint.x, adjustedEndPoint.y + bufferSize + MARKER_WIDTH],
    dagDirection === DagDirection.DOWN
      ? [adjustedEndPoint.x, adjustedEndPoint.y + MARKER_LENGTH]
      : [adjustedEndPoint.x + MARKER_LENGTH, adjustedEndPoint.y],
    dagDirection === DagDirection.DOWN
      ? [adjustedEndPoint.x + bufferSize + MARKER_WIDTH, adjustedEndPoint.y]
      : [adjustedEndPoint.x, adjustedEndPoint.y - bufferSize - MARKER_WIDTH],
    dagDirection === DagDirection.DOWN
      ? [adjustedEndPoint.x + bufferSize, adjustedEndPoint.y]
      : [adjustedEndPoint.x, adjustedEndPoint.y - bufferSize],
  ];

  // We need to add the buffer to the starting and ending points
  // taking DAG direction into account.
  const mask: [number, number][] = [
    dagDirection === DagDirection.DOWN
      ? [startPoint.x - bufferSize, startPoint.y]
      : [startPoint.x, startPoint.y + bufferSize],
    ...leftPath,
    ...endMarker,
    ...rightPath.reverse(),
    dagDirection === DagDirection.DOWN
      ? [startPoint.x + bufferSize, startPoint.y]
      : [startPoint.x, startPoint.y - bufferSize],
  ];

  return line()(mask) || '';
};

// It is possible to get the length of an SVG path using the
// built in getTotalLength function. With the current setup
// it is easier for us to calculate the length of the line manually.
const getLength = (arr: [number, number][]) => {
  const total = arr.reduce((acc, cur, index, array) => {
    let len = 0;
    if (index !== 0) {
      len =
        Math.abs(array[index - 1][0] - cur[0]) +
        Math.abs(array[index - 1][1] - cur[1]);
    }
    return acc + len;
  }, 0);

  return total;
};

const useLink = (
  link: Link,
  dagDirection: DagDirection,
  preorder?: string[],
  reversePreorder?: string[],
) => {
  const {hoveredNode} = useHoveredNode();
  const id = `${link.source.id}_${link.target.id}`;

  const displayLineArray = getLineArray(
    link.startPoint,
    link.endPoint,
    link.bendPoints,
    dagDirection,
  );

  const displayPath = useMemo(
    () => line()(displayLineArray) || '',
    [displayLineArray],
  );

  const animationLineArray = getLineArray(
    dagDirection === DagDirection.DOWN
      ? {x: link.startPoint.x, y: link.startPoint.y - CIRCLE_RADIUS * 2}
      : {x: link.startPoint.x - CIRCLE_RADIUS * 2, y: link.startPoint.y},
    dagDirection === DagDirection.DOWN
      ? {x: link.endPoint.x, y: link.endPoint.y + CIRCLE_RADIUS * 2}
      : {x: link.endPoint.x + CIRCLE_RADIUS * 2, y: link.endPoint.y},
    link.bendPoints,
    dagDirection,
  );
  const animationPath = useMemo(
    () => line()(animationLineArray) || '',
    [animationLineArray],
  );
  const clipPath = useMemo(
    () =>
      getLineClipPath(
        link.startPoint,
        link.endPoint,
        link.bendPoints,
        dagDirection,
      ),
    [dagDirection, link.bendPoints, link.endPoint, link.startPoint],
  );

  // length / desired speed = time
  const speed = useMemo(() => {
    const length = getLength(animationLineArray);
    return `${length / DESIRED_SPEED}s`;
  }, [animationLineArray]);

  const subDagSelected = useMemo(
    () =>
      (preorder &&
        preorder.includes(link.source.id) &&
        preorder.includes(link.target.id)) ||
      (reversePreorder &&
        reversePreorder.includes(link.source.id) &&
        reversePreorder.includes(link.target.id)),
    [link.source.id, link.target.id, preorder, reversePreorder],
  );

  const sourceOrTargetHovered =
    hoveredNode === link.source.id || hoveredNode === link.target.id;

  const {setHoveredNode} = useHoveredNode();

  const onMouseOver = useCallback(() => {
    select(`#link_${link.id}`).raise();
    setHoveredNode(link.id);
  }, [link.id, setHoveredNode]);

  const onMouseOut = useCallback(() => {
    setHoveredNode('');
  }, [setHoveredNode]);

  return {
    transferring: link.transferring,
    highlightLink: subDagSelected || sourceOrTargetHovered,
    isCrossProject: link.isCrossProject,
    speed,
    id,
    displayPath,
    animationPath,
    clipPath,
    onMouseOver,
    onMouseOut,
  };
};

export default useLink;
