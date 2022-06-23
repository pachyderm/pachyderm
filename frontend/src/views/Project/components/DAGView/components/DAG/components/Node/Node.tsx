import {NodeState, NodeType} from '@graphqlTypes';
import {SuccessCheckmark} from '@pachyderm/components';
import classNames from 'classnames';
import React from 'react';

import {Node as GraphQLNode} from '@dash-frontend/lib/types';

import NodeTooltip from './components/NodeTooltip';
import useNode from './hooks/useNode';
import styles from './Node.module.css';

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  nodeWidth: number;
  nodeHeight: number;
};

const NODE_ICON_X_OFFSET = 28;
const NODE_ICON_Y_OFFSET = 8;
const SELECTION_X_OFFSET = 6;
const EGRESS_NODE_IMAGE_X_OFFSET = 70;
const EGRESS_NODE_IMAGE_Y_OFFSET = -30;
const SUCCESS_CHECKMARK_X_OFFSET = 155;
const SUCCESS_CHECKMARK_Y_OFFSET = 10;
const INPUT_REPO_ICON_X_OFFSET = 10;
const INPUT_REPO_ICON_Y_OFFSET = 8;

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
    showSuccess,
    nodeIconHref,
    viewState,
  } = useNode(node, isInteractive);

  const classes = classNames(styles.nodeGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: selectedNode === node.name,
    [styles.hover]: isHovered,
    [styles.access]: node.access,
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
      {!isEgress && (
        <>
          <rect
            width={nodeWidth}
            height={nodeHeight}
            className={classNames(styles.node, {
              [styles.selected]: selectedNode === node.name,
              [styles.error]: node.state === NodeState.ERROR,
            })}
            rx={3}
            ry={3}
          />
          <text
            fontSize="14px"
            fontWeight="600"
            textAnchor="start"
            dominantBaseline="middle"
            className="nodeLabel"
            x={36}
            y={nodeHeight / 2}
          />
          {selectedNode === node.name && (
            <image
              x={nodeWidth - SELECTION_X_OFFSET}
              href={`/dag_node_selected${
                node.state === NodeState.ERROR ? '_error' : ''
              }.svg`}
            />
          )}
          <image
            x={nodeWidth - NODE_ICON_X_OFFSET}
            y={NODE_ICON_Y_OFFSET}
            pointerEvents="none"
            href={nodeIconHref}
            data-testid={`Node__state-${node.state}`}
          />
          {viewState.globalIdFilter && (
            <image
              x={nodeWidth - NODE_ICON_X_OFFSET}
              y={nodeHeight - NODE_ICON_Y_OFFSET - 16}
              pointerEvents="none"
              href="/dag_globalId.svg"
              data-testid="Node__globalId"
            />
          )}
        </>
      )}

      <image
        x={isEgress ? EGRESS_NODE_IMAGE_X_OFFSET : 0}
        y={isEgress ? EGRESS_NODE_IMAGE_Y_OFFSET : 0}
        href={getNodeImageHref(node)}
      />

      {node.type === NodeType.INPUT_REPO && (
        <image
          x={INPUT_REPO_ICON_X_OFFSET}
          y={INPUT_REPO_ICON_Y_OFFSET}
          pointerEvents="none"
          href="/dag_input_repo.svg"
          transform="scale(0.9)"
        />
      )}

      <SuccessCheckmark
        show={showSuccess}
        x={SUCCESS_CHECKMARK_X_OFFSET}
        y={SUCCESS_CHECKMARK_Y_OFFSET}
      />
      {NodeType.EGRESS === node.type && (
        <NodeTooltip node={node} show={isHovered} />
      )}
    </g>
  );
};

export default Node;
