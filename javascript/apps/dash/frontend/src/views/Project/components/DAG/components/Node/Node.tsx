import {SuccessCheckmark} from '@pachyderm/components';
import classNames from 'classnames';
import React from 'react';

import {Node as GraphQLNode, NodeType} from '@graphqlTypes';

import LeaveJobButton from './components/LeaveJobButton';
import NodeTooltip from './components/NodeTooltip';
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
}) => {
  const {
    isHovered,
    onClick,
    onMouseOut,
    onMouseOver,
    selectedNode,
    groupName,
    isEgress,
    normalizedNodeName,
    showSuccess,
    showLeaveJob,
    handleLeaveJobClick,
    closeLeaveJob,
    nodeIconHref,
  } = useNode(node, isInteractive);

  const classes = classNames(styles.nodeGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: selectedNode === node.name || isHovered,
    [styles.access]: node.access,
    [styles.showLeaveJob]: showLeaveJob,
  });

  const getNodeImageHref = (node: GraphQLNode) => {
    switch (node.type) {
      case NodeType.OUTPUT_REPO:
      case NodeType.INPUT_REPO:
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

      {showLeaveJob ? (
        <LeaveJobButton
          onClick={handleLeaveJobClick}
          onClose={closeLeaveJob}
          isRepo={
            node.type === NodeType.OUTPUT_REPO ||
            node.type === NodeType.INPUT_REPO
          }
        />
      ) : (
        <NodeTooltip node={node} show={isHovered && isInteractive} />
      )}

      <image
        x={isEgress ? EGRESS_NODE_IMAGE_X_OFFSET : NODE_IMAGE_X_OFFSET}
        y={isEgress ? EGRESS_NODE_IMAGE_Y_OFFSET : NODE_IMAGE_Y_OFFSET}
        pointerEvents="none"
        transform={`scale(${nodeHeight / ORIGINAL_NODE_IMAGE_HEIGHT})`}
        href={getNodeImageHref(node)}
      />

      {node.type === NodeType.OUTPUT_REPO && (
        <image
          x={NODE_IMAGE_X_OFFSET - 7}
          y={NODE_IMAGE_Y_OFFSET + 20}
          pointerEvents="none"
          href="/dag_output_plus.svg"
        />
      )}

      {!isEgress && (
        <image
          x={nodeWidth - NODE_ICON_X_OFFSET}
          y={NODE_ICON_Y_OFFSET}
          pointerEvents="none"
          href={nodeIconHref}
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
