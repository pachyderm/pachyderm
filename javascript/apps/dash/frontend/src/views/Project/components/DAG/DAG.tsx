import React from 'react';

import {Dag, DagDirection} from '@graphqlTypes';

import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';

import Link from './components/Link';
import Node from './components/Node';
import styles from './DAG.module.css';
import useDag from './hooks/useDag';

type DagProps = {
  dagsToShow: number;
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  rotateDag: () => void;
  dagDirection: DagDirection;
};

const DAG: React.FC<DagProps> = ({
  data,
  id,
  nodeWidth,
  nodeHeight,
  isInteractive = true,
  dagDirection,
}) => {
  const {handleRectClick, rectBox} = useDag({
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
        onClick={handleRectClick}
      />
      {/* Ordering of links and nodes in DOM is important so nodes are on top layer */}
      {data.links.map((link) => (
        <Link key={link.id} link={link} isInteractive={isInteractive} />
      ))}
      {data.nodes.map((node) => (
        <Node
          key={node.name}
          node={node}
          nodeHeight={NODE_HEIGHT}
          nodeWidth={NODE_WIDTH}
          isInteractive={isInteractive}
        />
      ))}
    </g>
  );
};

export default DAG;
