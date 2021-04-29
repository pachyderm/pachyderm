import {Group, Circle, Tooltip} from '@pachyderm/components';
import classnames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';
import {Dag} from '@graphqlTypes';

import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';

import Link from './components/Link';
import Node from './components/Node';
import styles from './DAG.module.css';
import useDag from './hooks/useDag';

type DagProps = {
  count: number;
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  setLargestDagScale: React.Dispatch<React.SetStateAction<number | null>>;
  largestDagScale: number | null;
};

const MARKERS = [
  {id: 'end-arrow', color: '#000'},
  {id: 'end-arrow-active', color: '#5ba3b1'},
  {id: 'end-arrow-error', color: '#E02020'},
];

const DAG: React.FC<DagProps> = ({
  count,
  data,
  id,
  nodeWidth,
  nodeHeight,
  isInteractive = true,
  setLargestDagScale,
  largestDagScale,
}) => {
  const {navigateToDag, svgSize} = useDag({
    count,
    data,
    id,
    isInteractive,
    largestDagScale,
    nodeHeight,
    nodeWidth,
    setLargestDagScale,
  });

  return (
    <HoveredNodeProvider>
      <div
        id={`${id}Base`}
        className={classnames(styles.base, {
          [styles.draggable]: isInteractive,
        })}
        onClick={!isInteractive ? () => navigateToDag(data.id) : noop}
      >
        {data.priorityPipelineState && (
          <Tooltip tooltipText={data.priorityPipelineState} tooltipKey="status">
            <span className={styles.pipelineStatus}>
              <Group spacing={8} align="center">
                <Circle color="red" />
                {readablePipelineState(data.priorityPipelineState)}
              </Group>
            </span>
          </Tooltip>
        )}
        <svg
          id={id}
          preserveAspectRatio="xMinYMid meet"
          viewBox={`0 0 ${svgSize.width} ${svgSize.height}`}
          className={styles.parent}
        >
          <defs>
            {MARKERS.map((marker) => (
              <marker
                key={marker.id}
                viewBox="0 -5 10 10"
                refX={9}
                markerWidth={5}
                markerHeight={5}
                orient="auto"
                id={marker.id}
              >
                <path d="M0,-5L10,0L0,5" fill={marker.color} />
              </marker>
            ))}
          </defs>
          <g id={`${id}Graph`}>
            {data.nodes.map((node) => (
              <Node
                key={node.name}
                node={node}
                nodeHeight={NODE_HEIGHT}
                nodeWidth={NODE_WIDTH}
                isInteractive={isInteractive}
              />
            ))}
            {data.links.map((link) => (
              <Link key={link.id} link={link} isInteractive={isInteractive} />
            ))}
          </g>
        </svg>
      </div>
    </HoveredNodeProvider>
  );
};

export default DAG;
