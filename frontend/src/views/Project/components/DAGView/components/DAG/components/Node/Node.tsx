import {NodeState, NodeType} from '@graphqlTypes';
import {
  SuccessCheckmark,
  RepoSVG,
  PipelineSVG,
  LockSVG,
  JobsSVG,
  StatusPausedSVG,
  StatusBusySVG,
  StatusWarningSVG,
  StatusRunningSVG,
  StatusCheckmarkSVG,
  ChevronRightSVG,
} from '@pachyderm/components';
import classNames from 'classnames';
import React from 'react';

import {Node as GraphQLNode} from '@dash-frontend/lib/types';
import {NODE_INPUT_REPO} from '@dash-frontend/views/Project/constants/nodeSizes';

import NodeTooltip from './components/NodeTooltip';
import {ReactComponent as EgressSVG} from './DagEgress.svg';
import useNode from './hooks/useNode';
import styles from './Node.module.css';

type NodeIconProps = {
  state: GraphQLNode['state'];
  x?: number;
  y?: number;
  width?: number;
  height?: number;
};

const NodeStateIcon = ({state, ...rest}: NodeIconProps) => {
  switch (state) {
    case NodeState.ERROR:
      return <StatusWarningSVG {...rest} />;
    case NodeState.RUNNING:
      return <StatusRunningSVG {...rest} />;
    case NodeState.BUSY:
      return <StatusBusySVG {...rest} />;
    case NodeState.PAUSED:
      return <StatusPausedSVG {...rest} />;
    case NodeState.SUCCESS:
      return <StatusCheckmarkSVG {...rest} />;
    default:
      return null;
  }
};

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  nodeWidth: number;
  nodeHeight: number;
};

const NODE_ICON_X_OFFSET = 90;
const NODE_ICON_Y_OFFSET = -8;
const BUTTON_HEIGHT = 48;
const BUTTON_WIDTH = 112;

const textElementProps = {
  fontSize: '14px',
  fontWeight: '600',
  textAnchor: 'start',
  dominantBaseline: 'middle',
  className: 'nodeLabel',
};

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
    repoSelected,
    pipelineSelected,
    groupName,
    isEgress,
    showSuccess,
  } = useNode(node, isInteractive);

  const pipelineClasses = classNames(styles.buttonGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: pipelineSelected,
    [styles.hover]: isHovered,
    [styles.access]: node.access,
  });

  const repoClasses = classNames(styles.buttonGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: repoSelected,
    [styles.hover]: isHovered,
    [styles.access]: node.access,
  });

  if (node.type === NodeType.INPUT_REPO) {
    return (
      <g
        className={repoClasses}
        id={groupName}
        transform={`translate (${node.x}, ${node.y})`}
        onClick={() => onClick('repo')}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
      >
        <rect
          className={styles.node}
          width={nodeWidth}
          height={NODE_INPUT_REPO}
          rx={3}
          ry={3}
        />
        <text {...textElementProps} />
        <g transform="scale(0.75)">
          {node.access ? <RepoSVG x={15} y={21} /> : <LockSVG x={15} y={23} />}
        </g>
      </g>
    );
  }

  if (isEgress) {
    return (
      <g
        id={groupName}
        transform={`translate (${node.x + 70}, ${node.y - 30})`}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
        onClick={() => onClick('pipeline')}
      >
        <SuccessCheckmark show={showSuccess} x={155} y={10} />
        {NodeType.EGRESS === node.type && (
          <NodeTooltip node={node} show={isHovered} />
        )}
        <EgressSVG />
      </g>
    );
  }

  const visiblePipelineStatus =
    node.state &&
    [NodeState.BUSY, NodeState.ERROR, NodeState.PAUSED].includes(node.state);

  return (
    <g id={groupName} transform={`translate (${node.x}, ${node.y})`}>
      <rect
        width={nodeWidth}
        height={nodeHeight}
        className={classNames(styles.node)}
        rx={3}
        ry={3}
      />
      <line
        x1="0"
        y1={nodeHeight - BUTTON_HEIGHT}
        x2={nodeWidth}
        y2={nodeHeight - BUTTON_HEIGHT}
        className={styles.line}
        stroke="black"
      />
      <text {...textElementProps} />

      {visiblePipelineStatus && (
        <g
          id="pipeineStatusGroup"
          data-testid={`Node__state-${node.state}`}
          transform={`translate (${
            nodeWidth - NODE_ICON_X_OFFSET - 8
          }, ${NODE_ICON_Y_OFFSET}) scale(0.6)`}
        >
          <rect
            width={44 / 0.6}
            height={19 / 0.6}
            className={styles.statusRect}
            rx={8}
            ry={8}
          />
          <NodeStateIcon state={node.state} x={10} y={6} />
          <PipelineSVG x={42} y={6} />
        </g>
      )}

      <g
        id="jobStatusGroup"
        data-testid={`Node__state-${node.jobState}`}
        transform={`translate (${
          nodeWidth - NODE_ICON_X_OFFSET / 2 - 4
        }, ${NODE_ICON_Y_OFFSET}) scale(0.6)`}
      >
        <rect
          width={44 / 0.6}
          height={19 / 0.6}
          className={styles.statusRect}
          rx={8}
          ry={8}
        />
        <NodeStateIcon state={node.jobState} x={10} y={6} />
        <JobsSVG x={42} y={6} />
      </g>

      <g
        id="pipelineButtonGroup"
        transform={`translate (0, ${nodeHeight - BUTTON_HEIGHT})`}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
        onClick={() => onClick('pipeline')}
        className={pipelineClasses}
      >
        <rect width={BUTTON_WIDTH} height={BUTTON_HEIGHT} rx={3} ry={3} />

        <g transform="scale(0.75)">
          {node.access ? (
            <PipelineSVG x={15} y={23} />
          ) : (
            <LockSVG x={15} y={23} />
          )}
        </g>
        <text {...textElementProps} x={32} y={26}>
          Pipeline
        </text>
      </g>

      <g transform={`scale(0.6)`}>
        <ChevronRightSVG
          x={nodeWidth / 2 / 0.6 - 10}
          y={nodeHeight / 0.6 - 28 / 0.6}
        />
      </g>

      <g
        id="repoButtonGroup"
        transform={`translate (${BUTTON_WIDTH + 20}, ${
          nodeHeight - BUTTON_HEIGHT
        })`}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
        onClick={() => onClick('repo')}
        className={repoClasses}
      >
        <rect width={BUTTON_WIDTH} height={BUTTON_HEIGHT} rx={3} ry={3} />

        <g transform="scale(0.75)">
          {node.access ? <RepoSVG x={15} y={23} /> : <LockSVG x={15} y={23} />}
        </g>
        <text {...textElementProps} x={32} y={26}>
          Output
        </text>
      </g>
    </g>
  );
};

export default Node;
