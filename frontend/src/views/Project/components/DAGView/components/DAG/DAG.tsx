import React from 'react';

import {Dags, DagDirection} from '@dash-frontend/lib/types';

import {NODE_HEIGHT, NODE_WIDTH} from '../../../../constants/nodeSizes';
import {SIDENAV_PADDING} from '../../hooks/useDAGView';

import Link from './components/Link';
import Node from './components/Node';
import styles from './DAG.module.css';
import useDag from './hooks/useDag';

const DAG_TOP_PADDING = 30;
const DAG_SIDE_PADDING = SIDENAV_PADDING;
const HIDE_DETAILS_THRESHOLD = 0.6;
const SIMPLE_DAG_THRESHOLD = 0.3;

type DagProps = {
  data: Dags | undefined;
  isInteractive?: boolean;
  rotateDag: () => void;
  dagDirection: DagDirection;
  largeDagMode?: boolean;
  forceFullRender?: boolean;
};

const DAG: React.FC<DagProps> = ({
  data,
  isInteractive = true,
  dagDirection,
  largeDagMode = false,
  forceFullRender = false,
}) => {
  const {translateX, translateY, scale, svgWidth, svgHeight} = useDag();

  return (
    <g className={styles.graph}>
      {/* Ordering of links and nodes in DOM is important so nodes are on top layer */}
      {data?.links.length !== 0 &&
        (!largeDagMode || forceFullRender || scale >= SIMPLE_DAG_THRESHOLD) &&
        data?.links.map((link) => {
          // only render links within DAG canvas boundaries
          const [linkMinX, linkMaxX] = [link.startPoint.x, link.endPoint.x]
            .sort((a, b) => a - b)
            .map((p) => p * scale + translateX + DAG_SIDE_PADDING);
          const [linkMinY, linkMaxY] = [link.startPoint.y, link.endPoint.y]
            .sort((a, b) => a - b)
            .map((p) => p * scale + translateY + DAG_TOP_PADDING);
          if (
            forceFullRender ||
            (linkMaxX > DAG_SIDE_PADDING &&
              linkMaxY > DAG_TOP_PADDING &&
              linkMinX < svgWidth + DAG_SIDE_PADDING &&
              linkMinY < svgHeight)
          ) {
            return (
              <Link key={link.id} link={link} dagDirection={dagDirection} />
            );
          }
          return null;
        })}
      {data?.nodes.length !== 0 &&
        data?.nodes.map((node) => {
          // only render nodes within DAG canvas boundaries
          const nodeRealX = node.x * scale + translateX + DAG_SIDE_PADDING;
          const nodeRealY = node.y * scale + DAG_TOP_PADDING + translateY;
          if (
            forceFullRender ||
            (nodeRealX > DAG_SIDE_PADDING - NODE_WIDTH * scale &&
              nodeRealY > DAG_TOP_PADDING - NODE_HEIGHT * scale &&
              nodeRealX < svgWidth + DAG_SIDE_PADDING &&
              nodeRealY < svgHeight + DAG_TOP_PADDING)
          ) {
            return (
              <Node
                key={node.id}
                node={node}
                isInteractive={
                  (!largeDagMode || scale >= HIDE_DETAILS_THRESHOLD) &&
                  isInteractive
                }
                hideDetails={largeDagMode && scale < HIDE_DETAILS_THRESHOLD}
                showSimple={largeDagMode && scale < SIMPLE_DAG_THRESHOLD}
              />
            );
          }
          return null;
        })}
    </g>
  );
};

export default DAG;
