import {SuccessCheckmark} from '@pachyderm/components';
import classNames from 'classnames';
import React from 'react';

import {Node as GraphQLNode, NodeType} from '@graphqlTypes';

import NodeTooltip from './components/NodeTooltip';
import useNode from './hooks/useNode';
import styles from './Node.module.css';

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  nodeWidth: number;
  nodeHeight: number;
  offset: {x: number; y: number};
};

const NODE_ICON_X_OFFSET = 12;
const NODE_ICON_Y_OFFSET = -8;
const NODE_IMAGE_Y_OFFSET = -40;
const NODE_IMAGE_X_OFFSET = 76;
const EGRESS_NODE_IMAGE_X_OFFSET = 40;
const EGRESS_NODE_IMAGE_Y_OFFSET = -20;
const ORIGINAL_NODE_IMAGE_HEIGHT = 102;
const SUCCESS_CHECKMARK_X_OFFSET = 65;
const SUCCESS_CHECKMARK_Y_OFFSET = 10;

const Node: React.FC<NodeProps> = ({
  node,
  isInteractive,
  nodeHeight,
  nodeWidth,
  offset,
}) => {
  const {
    hoveredNode,
    onClick,
    onMouseOut,
    onMouseOver,
    selectedNode,
    state,
    groupName,
    isEgress,
    normalizedNodeName,
    showSuccess,
  } = useNode(node, isInteractive, offset);

  const classes = classNames(styles.nodeGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: [selectedNode, hoveredNode].includes(node.name),
    [styles.access]: node.access,
  });

  const getNodeIconHref = (state: string, access: boolean) => {
    if (!access) {
      return '/dag_no_access.svg';
    }

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
    switch (node.type) {
      case NodeType.REPO:
        return '/dag_repo.svg';
      case NodeType.PIPELINE:
        return '/dag_pipeline.svg';
      case NodeType.EGRESS:
        return '/dag_egress.svg';
    }
  };

  return (
    <g
      className={classes}
      id={groupName}
      transform={`translate (${node.x + offset.x}, ${node.y + offset.y})`}
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
          <p
            className={classNames(styles.label, {
              [styles.egress]: isEgress,
            })}
            style={{height: `${nodeHeight}px`}}
          >
            {normalizedNodeName}
          </p>
        </span>
      </foreignObject>

      <NodeTooltip node={node} textClassName={styles.tooltipText} />

      <image
        x={isEgress ? EGRESS_NODE_IMAGE_X_OFFSET : NODE_IMAGE_X_OFFSET}
        y={isEgress ? EGRESS_NODE_IMAGE_Y_OFFSET : NODE_IMAGE_Y_OFFSET}
        pointerEvents="none"
        transform={`scale(${nodeHeight / ORIGINAL_NODE_IMAGE_HEIGHT})`}
        href={getNodeImageHref(node)}
      />

      {!isEgress && (
        <image
          x={nodeWidth - NODE_ICON_X_OFFSET}
          y={NODE_ICON_Y_OFFSET}
          pointerEvents="none"
          href={getNodeIconHref(state, node.access)}
        />
      )}

      <SuccessCheckmark
        show={showSuccess}
        x={SUCCESS_CHECKMARK_X_OFFSET}
        y={SUCCESS_CHECKMARK_Y_OFFSET}
      />
    </g>
  );
};

export default Node;
