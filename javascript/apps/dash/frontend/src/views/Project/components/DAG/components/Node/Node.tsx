import classNames from 'classnames';
import React from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Node as GraphQLNode, NodeType} from '@graphqlTypes';

import deriveRepoNameFromNode from '../../utils/deriveRepoNameFromNode';

import useNode from './hooks/useNode';
import styles from './Node.module.css';

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  nodeWidth: number;
  nodeHeight: number;
};

const NODE_ICON_X_OFFSET = 12;
const NODE_ICON_Y_OFFSET = -8;
const NODE_IMAGE_Y_OFFSET = -40;
const NODE_IMAGE_X_OFFSET = 76;
const ORIGINAL_NODE_IMAGE_HEIGHT = 102;
const NODE_TOOLTIP_WIDTH = 250;
const NODE_TOOLTIP_HEIGHT = 80;
const NODE_TOOLTIP_OFFSET = -92;

const Node: React.FC<NodeProps> = ({
  node,
  isInteractive,
  nodeHeight,
  nodeWidth,
}) => {
  const {
    hoveredNode,
    onClick,
    onMouseOut,
    onMouseOver,
    selectedNode,
    state,
    groupName,
  } = useNode(node, isInteractive);

  const classes = classNames(styles.nodeGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: [selectedNode, hoveredNode].includes(node.name),
  });

  const getNodeIconHref = (state: string) => {
    switch (state) {
      case 'busy':
        return '/dag_busy.svg';
      case 'error':
        return '/dag_pipeline_error.svg';
      case 'paused':
        return '/dag_paused.svg';
    }
  };

  const getNodeImageHref = (node: GraphQLNode) => {
    if (!node.access) {
      return '/dag_no_access.svg';
    }
    switch (node.type) {
      case NodeType.REPO:
        return '/dag_repo.svg';
      case NodeType.PIPELINE:
        return '/dag_pipeline.svg';
    }
  };

  return (
    <g
      className={classes}
      id={groupName}
      transform={`translate (${node.x}, ${node.y})`}
      onClick={onClick}
      onMouseOver={onMouseOver}
      onMouseOut={onMouseOut}
    >
      <foreignObject
        className={styles.node}
        width={nodeWidth}
        height={nodeHeight}
      >
        <span>
          <p className={styles.label} style={{height: `${nodeHeight}px`}}>
            {node.access ? deriveRepoNameFromNode(node) : 'No Access'}
          </p>
        </span>
      </foreignObject>

      <foreignObject
        className={styles.nodeTooltip}
        width={NODE_TOOLTIP_WIDTH}
        height={NODE_TOOLTIP_HEIGHT}
        y={NODE_TOOLTIP_OFFSET}
        x={-(NODE_TOOLTIP_WIDTH / 4)}
      >
        <span>
          <p
            className={styles.tooltipText}
            style={{minHeight: `${NODE_TOOLTIP_HEIGHT - 9}px`}}
          >
            {deriveRepoNameFromNode(node)}
            <br />
            <br />
            {node.type === NodeType.PIPELINE
              ? `${node.type.toLowerCase()} status: ${readablePipelineState(
                  node.state || '',
                )}`
              : ''}
          </p>
        </span>
      </foreignObject>

      <image
        x={NODE_IMAGE_X_OFFSET}
        y={NODE_IMAGE_Y_OFFSET}
        pointerEvents="none"
        transform={`scale(${nodeHeight / ORIGINAL_NODE_IMAGE_HEIGHT})`}
        href={getNodeImageHref(node)}
      />

      <image
        x={nodeWidth - NODE_ICON_X_OFFSET}
        y={NODE_ICON_Y_OFFSET}
        pointerEvents="none"
        href={getNodeIconHref(state)}
      />
    </g>
  );
};

export default Node;
