import classNames from 'classnames';
import React from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Node as GraphQLNode, NodeType} from '@graphqlTypes';

import deriveRepoNameFromNode from '../../utils/deriveRepoNameFromNode';

import useNode from './hooks/useNode';
import {ReactComponent as Busy} from './images/busy.svg';
import {ReactComponent as Error} from './images/error.svg';
import {ReactComponent as NoAccess} from './images/noAccess.svg';
import {ReactComponent as Pipeline} from './images/pipeline.svg';
import {ReactComponent as Repo} from './images/repo.svg';
import styles from './Node.module.css';

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  nodeWidth: number;
  nodeHeight: number;
};

const NODE_ICON_X_OFFSET = 20;
const NODE_ICON_Y_OFFSET = -15;
const NODE_IMAGE_Y_OFFSET = -20;
const NODE_IMAGE_PREVIEW_Y_OFFSET = -20;
const ORIGINAL_NODE_IMAGE_WIDTH = 145;
const ORIGINAL_NODE_IMAGE_HEIGHT = 102;
const NODE_TOOLTIP_WIDTH = 250;
const NODE_TOOLTIP_HEIGHT = 80;
const NODE_TOOLTIP_OFFSET = -88;

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
  } = useNode(node, isInteractive);

  const classes = classNames(styles.nodeGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: [selectedNode, hoveredNode].includes(node.name),
  });

  return (
    <g
      className={classes}
      id={`${node.name}GROUP`}
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
            {node.access ? node.name : 'No Access'}
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
      <g
        transform={`scale(${nodeHeight / ORIGINAL_NODE_IMAGE_HEIGHT})`}
        pointerEvents="none"
      >
        {node.access ? (
          node.type === NodeType.REPO ? (
            <Repo
              y={
                isInteractive
                  ? NODE_IMAGE_Y_OFFSET
                  : NODE_IMAGE_PREVIEW_Y_OFFSET
              }
              x={ORIGINAL_NODE_IMAGE_WIDTH - nodeWidth}
            />
          ) : (
            <Pipeline
              y={
                isInteractive
                  ? NODE_IMAGE_Y_OFFSET
                  : NODE_IMAGE_PREVIEW_Y_OFFSET
              }
              x={ORIGINAL_NODE_IMAGE_WIDTH - nodeWidth}
            />
          )
        ) : (
          <NoAccess
            y={
              isInteractive ? NODE_IMAGE_Y_OFFSET : NODE_IMAGE_PREVIEW_Y_OFFSET
            }
            x={ORIGINAL_NODE_IMAGE_WIDTH - nodeWidth}
          />
        )}
      </g>

      {state === 'busy' ? (
        <Busy
          x={nodeWidth - NODE_ICON_X_OFFSET}
          y={NODE_ICON_Y_OFFSET}
          pointerEvents="none"
        />
      ) : state === 'error' ? (
        <Error
          x={nodeWidth - NODE_ICON_X_OFFSET}
          y={NODE_ICON_Y_OFFSET}
          pointerEvents="none"
        />
      ) : null}
    </g>
  );
};

export default Node;
