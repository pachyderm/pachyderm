import {NodeState, NodeType} from '@graphqlTypes';
import classNames from 'classnames';
import React, {SVGProps} from 'react';

import {Node as GraphQLNode} from '@dash-frontend/lib/types';
import {
  NODE_INPUT_REPO,
  NODE_HEIGHT,
  NODE_WIDTH,
} from '@dash-frontend/views/Project/constants/nodeSizes';
import {
  SuccessCheckmark,
  RepoSVG,
  PipelineSVG,
  LockSVG,
  JobsSVG,
  StatusPausedSVG,
  StatusDotsSVG,
  StatusWarningSVG,
  StatusCheckmarkSVG,
  ChevronRightSVG,
} from '@pachyderm/components';

import NodeTooltip from './components/NodeTooltip';
import {ReactComponent as EgressSVG} from './DagEgress.svg';
import useNode from './hooks/useNode';
import styles from './Node.module.css';

interface NodeIconProps extends SVGProps<SVGSVGElement> {
  state: GraphQLNode['nodeState'];
}

const NodeStateIcon = ({state, ...rest}: NodeIconProps) => {
  switch (state) {
    case NodeState.ERROR:
      return <StatusWarningSVG color="var(--pachyderm-red)" {...rest} />;
    case NodeState.RUNNING:
      return <StatusDotsSVG color="var(--icon-green)" {...rest} />;
    case NodeState.BUSY:
      return <StatusDotsSVG color="var(--pachyderm-yellow)" {...rest} />;
    case NodeState.PAUSED:
      return <StatusPausedSVG color="var(--pachyderm-yellow)" {...rest} />;
    case NodeState.SUCCESS:
      return <StatusCheckmarkSVG color="var(--icon-green)" {...rest} />;
    default:
      return null;
  }
};

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  hideDetails?: boolean;
  showSimple?: boolean;
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
  hideDetails = false,
  showSimple = false,
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
  } = useNode(node, isInteractive, hideDetails);

  const pipelineClasses = classNames(styles.buttonGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: pipelineSelected,
    [styles.access]: node.access,
  });

  const repoClasses = classNames(styles.buttonGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: repoSelected,
    [styles.access]: node.access,
  });

  const statusClasses = classNames(styles.statusRect, {
    [styles.interactive]: isInteractive,
    [styles.access]: node.access,
  });

  if (node.type === NodeType.INPUT_REPO) {
    return (
      <g
        role="button"
        aria-label={`${groupName} repo`}
        className={repoClasses}
        id={groupName}
        transform={`translate (${node.x}, ${node.y})`}
        onClick={() => onClick('repo')}
      >
        <rect
          className={classNames(styles.node, {
            [styles.repoSimplifiedBox]: showSimple,
          })}
          width={NODE_WIDTH}
          height={NODE_INPUT_REPO}
          rx={showSimple ? 15 : 3}
          ry={showSimple ? 15 : 3}
        />
        {!hideDetails && (
          <>
            <text {...textElementProps} />
            <g transform="scale(0.75)">
              {node.access ? (
                <RepoSVG x={15} y={21} />
              ) : (
                <LockSVG color="var(--disabled-tertiary)" x={15} y={23} />
              )}
            </g>
          </>
        )}
      </g>
    );
  }

  if (isEgress) {
    return (
      <g
        role="button"
        aria-label={`${groupName} egress`}
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
    node.nodeState &&
    [NodeState.BUSY, NodeState.ERROR, NodeState.PAUSED].includes(
      node.nodeState,
    );

  return (
    <g id={groupName} transform={`translate (${node.x}, ${node.y})`}>
      <rect
        width={NODE_WIDTH}
        height={NODE_HEIGHT}
        className={classNames(styles.node, {
          [styles.pipelineSimplifiedBox]:
            showSimple && node.jobNodeState !== NodeState.ERROR,
          [styles.pipelineSimplifiedBoxError]:
            showSimple && node.jobNodeState === NodeState.ERROR,
        })}
        rx={showSimple ? 20 : 3}
        ry={showSimple ? 20 : 3}
      />
      {hideDetails && !showSimple && node.jobNodeState === NodeState.ERROR && (
        <g transform="scale (1.75)">
          <NodeStateIcon state={node.jobNodeState} x={45} y={17} />
          <JobsSVG x={70} y={17} />
        </g>
      )}
      {!hideDetails && (
        <>
          <text {...textElementProps} />
          <line
            x1="0"
            y1={NODE_HEIGHT - BUTTON_HEIGHT}
            x2={NODE_WIDTH}
            y2={NODE_HEIGHT - BUTTON_HEIGHT}
            className={styles.line}
            stroke="black"
          />
          {visiblePipelineStatus && (
            <g
              role="button"
              aria-label={`${groupName} status`}
              id="pipelineStatusGroup"
              data-testid={`Node__state-${node.nodeState}`}
              transform={`translate (${
                NODE_WIDTH - NODE_ICON_X_OFFSET - 8
              }, ${NODE_ICON_Y_OFFSET}) scale(0.6)`}
              onClick={() => onClick('status')}
              className={styles.statusGroup}
            >
              <rect
                width={44 / 0.6}
                height={19 / 0.6}
                className={statusClasses}
                rx={8}
                ry={8}
              />
              <NodeStateIcon state={node.nodeState} x={10} y={6} />
              <PipelineSVG x={42} y={6} />
            </g>
          )}

          {node.jobNodeState !== NodeState.IDLE && (
            <g
              role="button"
              aria-label={`${groupName} logs`}
              id="jobStatusGroup"
              data-testid={`Node__state-${node.jobNodeState}`}
              transform={`translate (${
                NODE_WIDTH - NODE_ICON_X_OFFSET / 2 - 4
              }, ${NODE_ICON_Y_OFFSET}) scale(0.6)`}
              onClick={() => onClick('logs')}
              className={styles.statusGroup}
            >
              <rect
                width={44 / 0.6}
                height={19 / 0.6}
                className={statusClasses}
                rx={8}
                ry={8}
              />
              <NodeStateIcon state={node.jobNodeState} x={10} y={6} />
              <JobsSVG x={42} y={6} />
            </g>
          )}
          <g
            role="button"
            aria-label={`${groupName} pipeline`}
            id="pipelineButtonGroup"
            transform={`translate (0, ${NODE_HEIGHT - BUTTON_HEIGHT})`}
            onClick={() => onClick('pipeline')}
            className={pipelineClasses}
          >
            <rect width={BUTTON_WIDTH} height={BUTTON_HEIGHT} rx={3} ry={3} />

            <g transform="scale(0.75)">
              {node.access ? (
                <PipelineSVG x={15} y={23} />
              ) : (
                <LockSVG color="var(--disabled-tertiary)" x={15} y={23} />
              )}
            </g>
            <text {...textElementProps} x={32} y={26}>
              Pipeline
            </text>
          </g>

          <g transform={`scale(0.6)`}>
            <ChevronRightSVG
              x={NODE_WIDTH / 2 / 0.6 - 10}
              y={NODE_HEIGHT / 0.6 - 28 / 0.6}
            />
          </g>

          <g
            role="button"
            aria-label={`${groupName} repo`}
            id="repoButtonGroup"
            transform={`translate (${BUTTON_WIDTH + 20}, ${
              NODE_HEIGHT - BUTTON_HEIGHT
            })`}
            onClick={() => onClick('repo')}
            className={repoClasses}
          >
            <rect width={BUTTON_WIDTH} height={BUTTON_HEIGHT} rx={3} ry={3} />

            <g transform="scale(0.75)">
              {node.access ? (
                <RepoSVG x={15} y={23} />
              ) : (
                <LockSVG color="var(--disabled-tertiary)" x={15} y={23} />
              )}
            </g>
            <text {...textElementProps} x={32} y={26}>
              Output
            </text>
          </g>
        </>
      )}
    </g>
  );
};

export default Node;
