import classnames from 'classnames';
import React from 'react';

import {DagDirection, Dags} from '@dash-frontend/lib/types';

import {
  BOTTOM_ROUNDED_BUTTON_PATH,
  CONNECTED_EGRESS_GRADIENT_MASK,
  NODE_HEIGHT,
  NODE_WIDTH,
  PIPELINE_GRADIENT_MASK,
  REPO_GRADIENT_MASK,
  TOP_ROUNDED_BUTTON_PATH,
} from '../../../../constants/nodeSizes';
import {SIDENAV_PADDING} from '../../hooks/useDAGView';
import HoveredNodeProvider from '../../providers/HoveredNodeProvider';
import HoverStats from '../HoverStats';

import Link from './components/Link';
import Node from './components/Node';
import styles from './DAG.module.css';
import useDag from './hooks/useDag';

const DAG_TOP_PADDING = 30;
const DAG_SIDE_PADDING = SIDENAV_PADDING;
const HIDE_DETAILS_THRESHOLD = 0.6;
const SIMPLE_DAG_THRESHOLD = 0.3;

const LARGE_DAG_MIN = 50;

const MARKERS = [
  {id: 'end-arrow-idle', color: '#A4A8AE'},
  {id: 'end-arrow-transferring', color: '#9BD49D'},
  {id: 'end-arrow-hover', color: '#272728'},
  {id: 'end-arrow-highlighted', color: '#99CAFF'},
];

type DagProps = {
  data: Dags | undefined;
  svgSize: {
    height: number;
    width: number;
  };
  rotateDag: () => void;
  forceFullRender?: boolean;
  dagDirection: DagDirection;
};

const DAG: React.FC<DagProps> = ({
  data,
  svgSize,
  forceFullRender = false,
  dagDirection,
}) => {
  const {
    translateX,
    translateY,
    scale,
    svgWidth,
    svgHeight,
    selectedNodePreorder,
    selectedNodeReversePreorder,
  } = useDag(data);

  const totalNodes = data?.nodes.length || 0;
  const largeDagMode = totalNodes > LARGE_DAG_MIN;

  const hideDetails = largeDagMode && scale < HIDE_DETAILS_THRESHOLD;
  const showSimple = largeDagMode && scale < SIMPLE_DAG_THRESHOLD;

  return (
    <HoveredNodeProvider>
      <HoverStats
        nodes={data?.nodes}
        links={data?.links}
        simplifiedView={hideDetails || showSimple}
      />

      <svg
        id="Svg"
        className={styles.base}
        preserveAspectRatio="xMinYMid meet"
        viewBox={`0 0 ${svgSize.width} ${svgSize.height}`}
      >
        <defs>
          {MARKERS.map((marker) => (
            <marker
              key={marker.id}
              viewBox="0 -5 10 10"
              refX={0}
              refY={0}
              markerWidth={5}
              markerHeight={5}
              orient="auto"
              id={marker.id}
            >
              <path d="M0,-5L10,0L0,5" fill={marker.color} />
            </marker>
          ))}
          <filter id="node-dropshadow">
            <feDropShadow dx="2" dy="2" stdDeviation="4" floodColor="#d6d6d7" />
          </filter>
          <linearGradient id="green-gradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#D9E7E7" />
            <stop offset="100%" stopColor="#005D5D" />
          </linearGradient>
          <linearGradient id="grey-gradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#EBEBEB26" />
            <stop offset="100%" stopColor="#EBEBEB" />
          </linearGradient>
          <linearGradient
            id="purple-gradient"
            x1="0%"
            y1="0%"
            x2="0%"
            y2="100%"
          >
            <stop offset="0%" stopColor="#6929C4" />
            <stop offset="100%" stopColor="#6929c426" />
          </linearGradient>
          <path id="topRoundedButton" d={TOP_ROUNDED_BUTTON_PATH} />
          <path id="bottomRoundedButton" d={BOTTOM_ROUNDED_BUTTON_PATH} />
          <clipPath id="topRoundedButtonOnly">
            <use xlinkHref="#topRoundedButton" />
          </clipPath>
          <clipPath id="bottomRoundedButtonOnly">
            <use xlinkHref="#bottomRoundedButton" />
          </clipPath>
          <path id="repoMask" d={REPO_GRADIENT_MASK} />
          <path id="connectedEgressMask" d={CONNECTED_EGRESS_GRADIENT_MASK} />
          <path id="pipelineMask" d={PIPELINE_GRADIENT_MASK} />
        </defs>
        <g id="Dags">
          <g
            className={classnames(styles.graph, {
              [styles.speed]:
                !forceFullRender &&
                largeDagMode &&
                scale < HIDE_DETAILS_THRESHOLD,
            })}
          >
            {/* Ordering of links and nodes in DOM is important so nodes are on top layer */}
            {data?.links.length !== 0 &&
              (!largeDagMode ||
                forceFullRender ||
                scale >= SIMPLE_DAG_THRESHOLD) &&
              data?.links.map((link) => {
                // only render links within DAG canvas boundaries
                const [linkMinX, linkMaxX] = [
                  link.startPoint.x,
                  link.endPoint.x,
                ]
                  .sort((a, b) => a - b)
                  .map((p) => p * scale + translateX + DAG_SIDE_PADDING);
                const [linkMinY, linkMaxY] = [
                  link.startPoint.y,
                  link.endPoint.y,
                ]
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
                    <Link
                      key={link.id}
                      link={link}
                      dagDirection={dagDirection}
                      preorder={selectedNodePreorder}
                      reversePreorder={selectedNodeReversePreorder}
                      hideDetails={hideDetails}
                    />
                  );
                }
                return null;
              })}
            {data?.nodes.length !== 0 &&
              data?.nodes.map((node) => {
                // only render nodes within DAG canvas boundaries
                const nodeRealX =
                  node.x * scale + translateX + DAG_SIDE_PADDING;
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
                        !largeDagMode || scale >= HIDE_DETAILS_THRESHOLD
                      }
                      hideDetails={hideDetails}
                      showSimple={showSimple}
                    />
                  );
                }
                return null;
              })}
          </g>
        </g>
      </svg>
    </HoveredNodeProvider>
  );
};

export default DAG;
