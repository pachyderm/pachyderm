import React from 'react';

import {Dag, DagDirection} from '@dash-frontend/lib/types';

import {NODE_HEIGHT, NODE_WIDTH} from '../../../../constants/nodeSizes';

import Link from './components/Link';
import Node from './components/Node';
import styles from './DAG.module.css';
import useDag from './hooks/useDag';

const DAG_TOP_PADDING = 30;
const DAG_SIDE_PADDING = 200;
type DagProps = {
  dagsToShow: number;
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  rotateDag: () => void;
  dagDirection: DagDirection;
  forceFullRender?: boolean;
};

const DAG: React.FC<DagProps> = ({
  data,
  id,
  nodeWidth,
  nodeHeight,
  isInteractive = true,
  dagDirection,
  forceFullRender = false,
}) => {
  const {rectBox, translateX, translateY, scale, svgWidth, svgHeight} = useDag({
    data,
    id,
    nodeHeight,
    nodeWidth,
    dagDirection,
  });

  return (
    <g id={id} className={styles.graph}>
      <rect
        className="borderRect"
        x={rectBox.x}
        y={rectBox.y}
        width={rectBox.width}
        height={rectBox.height}
      />
      {/* Ordering of links and nodes in DOM is important so nodes are on top layer */}
      {data.links.map((link) => {
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
          return <Link key={link.id} link={link} dagDirection={dagDirection} />;
        }
        return null;
      })}
      {data.nodes.map((node) => {
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
              nodeHeight={NODE_HEIGHT}
              nodeWidth={NODE_WIDTH}
              isInteractive={isInteractive}
            />
          );
        }
        return null;
      })}
    </g>
  );
};

export default DAG;
