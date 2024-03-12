import classNames from 'classnames';
import React, {SVGProps} from 'react';

import {readableJobState} from '@dash-frontend/lib/jobs';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {
  NodeState,
  NodeType,
  Node as GraphQLNode,
} from '@dash-frontend/lib/types';
import {
  NODE_INPUT_REPO,
  NODE_HEIGHT,
  NODE_WIDTH,
  CONNECTED_EGRESS_NODE_HEIGHT,
  BUTTON_MARGIN,
  BUTTON_WIDTH,
  BUTTON_HEIGHT,
  BORDER_RADIUS,
  STATUS_BUTTON_HEIGHT,
  STATUS_BUTTON_WIDTH,
  FOOTER_HEIGHT,
  BORDER_WIDTH,
} from '@dash-frontend/views/Project/constants/nodeSizes';
import {
  RepoSVG,
  PipelineSVG,
  JobsSVG,
  StatusPausedSVG,
  StatusWarningSVG,
  StatusCheckmarkSVG,
  SpinnerSVG,
  EgressSVG,
  LinkSVG,
  UploadSVG,
  PachProjectSVG,
  PachCronSVG,
  PachSpoutSVG,
  PachServiceSVG,
  PachSpoutAndServiceSVG,
  ParallelismSVG,
  CopySVG,
  CheckmarkSVG,
} from '@pachyderm/components';

import SVGButton from './components/SVGButton';
import useNode from './hooks/useNode';
import styles from './Node.module.css';

interface NodeIconProps extends SVGProps<SVGSVGElement> {
  state: GraphQLNode['nodeState'];
}

const SIMPLE_BORDER_RADIUS = 20;
const FOOTER_LEFT_MARGIN = 12;

const NodeStateIcon = ({state, ...rest}: NodeIconProps) => {
  switch (state) {
    case NodeState.ERROR:
      return <StatusWarningSVG color="var(--pachyderm-red)" {...rest} />;
    case NodeState.RUNNING:
    case NodeState.BUSY:
      return <SpinnerSVG {...rest} />;
    case NodeState.PAUSED:
      return <StatusPausedSVG color="var(--pachyderm-yellow)" {...rest} />;
    case NodeState.IDLE:
    case NodeState.SUCCESS:
      return <StatusCheckmarkSVG color="var(--icon-green)" {...rest} />;
    default:
      return null;
  }
};

const PipelineFooterIcon = ({type}: {type: NodeType}) => {
  switch (type) {
    case NodeType.SPOUT_SERVICE_PIPELINE:
      return <PachSpoutAndServiceSVG fill="#666" />;
    case NodeType.SPOUT_PIPELINE:
      return <PachSpoutSVG fill="#666" />;
    case NodeType.SERVICE_PIPELINE:
      return <PachServiceSVG fill="#666" />;
    default:
      return null;
  }
};

const pipelineFooterText = (type: NodeType) => {
  switch (type) {
    case NodeType.SPOUT_SERVICE_PIPELINE:
      return 'Spout and Service';
    case NodeType.SPOUT_PIPELINE:
      return 'Spout';
    case NodeType.SERVICE_PIPELINE:
      return 'Service';
    default:
      return null;
  }
};

const parallelismInfo = (id: string, parallelism?: string) => {
  if (parallelism) {
    const displayText = parseInt(parallelism) > 100 ? '>100' : parallelism;

    return (
      <g
        transform={` translate (${
          NODE_WIDTH - (displayText.length * 7 + 20)
        }, 12)
        scale(0.75)`}
        className={styles.tooltipWrapper}
        id={id}
      >
        <ParallelismSVG fill="#666" x="-25" y="-15" />
        <text fill="#666">{displayText}</text>
        <rect
          width={20 + (displayText.length * 5 + 20)}
          height="24"
          x="-28"
          y="-17"
          fill="transparent"
        >
          <title>
            {`${parallelism} maximum worker${
              parallelism !== '1' ? 's' : ''
            } based on the parallelism spec.`}
          </title>
        </rect>
      </g>
    );
  }
};

type NodeProps = {
  node: GraphQLNode;
  isInteractive: boolean;
  hideDetails?: boolean;
  showSimple?: boolean;
};

const labelTextStyle = {
  fontSize: '12px',
  fontWeight: '400',
  fontFamily: 'Montserrat',
  fill: '#666',
};

const Node: React.FC<NodeProps> = ({
  node,
  isInteractive,
  hideDetails = false,
  showSimple = false,
}) => {
  const {
    onClick,
    repoSelected,
    pipelineSelected,
    onMouseOut,
    onMouseOver,
    groupName,
    copied,
  } = useNode(node, isInteractive, hideDetails);

  const statusClasses = classNames(styles.statusGroup, {
    [styles.interactive]: isInteractive,
    [styles.access]: node.access,
  });

  const statusTextClasses = (state?: NodeState) =>
    classNames({
      [styles.errorText]: NodeState.ERROR === state,
      [styles.successText]:
        state && [NodeState.IDLE, NodeState.SUCCESS].includes(state),
    });

  if (node.type === NodeType.REPO || node.type === NodeType.CRON_REPO) {
    return (
      <g
        id={groupName}
        transform={`translate (${node.x}, ${node.y})`}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
      >
        {/* Node Container */}
        <rect
          width={NODE_WIDTH}
          height={NODE_INPUT_REPO}
          rx={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
          ry={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
          className={classNames(styles.node, {
            [styles.simplifiedInput]: showSimple,
            [styles.noShadow]: showSimple,
          })}
        />

        {/* Gradient - mask defined in DAGView.tsx */}
        {!showSimple && (
          <>
            {node.type === NodeType.CRON_REPO ? (
              <use
                xlinkHref="#repoMask"
                className={classNames(styles.purpleGradient, {
                  [styles.hideDetails]: hideDetails,
                })}
              />
            ) : (
              <use
                xlinkHref="#repoMask"
                className={classNames(styles.greenGradient, {
                  [styles.hideDetails]: hideDetails,
                })}
              />
            )}
          </>
        )}

        {!hideDetails && (
          <>
            {/* Top Button */}
            <SVGButton
              access={node.access}
              isInteractive
              isSelected={repoSelected}
              aria-label={`${groupName} repo`}
              onClick={() => onClick('repo')}
              transform={`translate (${BUTTON_MARGIN},${BUTTON_MARGIN})`}
              IconSVG={<RepoSVG x="11" y="11" />}
            />
            {/* Upload Button or CRON text*/}
            {node.type === NodeType.CRON_REPO ? (
              <g
                transform={`translate (${FOOTER_LEFT_MARGIN}, ${
                  NODE_INPUT_REPO - FOOTER_HEIGHT + FOOTER_HEIGHT / 4
                })`}
              >
                <g transform="scale(0.75)">
                  <PachCronSVG fill="#666" />
                </g>
                <text style={labelTextStyle} x="25" y="12">
                  CRON
                </text>
              </g>
            ) : (
              <g
                role="button"
                aria-label={`${groupName} file upload`}
                transform={`translate (${BUTTON_MARGIN}, ${
                  NODE_INPUT_REPO - FOOTER_HEIGHT + BUTTON_MARGIN
                })`}
                onClick={() => onClick('upload')}
                className={statusClasses}
              >
                <rect
                  width={BUTTON_WIDTH}
                  height={BUTTON_HEIGHT - BUTTON_MARGIN * 2}
                  rx={BORDER_RADIUS}
                  ry={BORDER_RADIUS}
                  className={styles.statusRect}
                />
                <g transform="translate (70, 5)">
                  <g transform="scale(0.75)">
                    <UploadSVG />
                  </g>
                  <text x="20" y="12">
                    Upload Files
                  </text>
                </g>
              </g>
            )}
          </>
        )}
      </g>
    );
  }

  if (node.type === NodeType.EGRESS) {
    return (
      <g
        aria-label={`${groupName} egress`}
        id={groupName}
        transform={`translate (${node.x}, ${node.y})`}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
      >
        {/* Container rectangle */}
        <rect
          width={NODE_WIDTH}
          height={CONNECTED_EGRESS_NODE_HEIGHT}
          rx={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
          ry={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
          className={classNames(styles.node, {
            [styles.simplifiedInput]: showSimple,
          })}
        />
        {!showSimple && (
          <>
            {/* Gradient - mask defined in DAGView.tsx */}
            <use
              xlinkHref="#connectedEgressMask"
              className={classNames(styles.greyGradient, {
                [styles.hideDetails]: hideDetails,
              })}
            />

            <rect
              width={NODE_WIDTH + BORDER_WIDTH}
              height={CONNECTED_EGRESS_NODE_HEIGHT + BORDER_WIDTH}
              transform="translate (-1, -1)"
              rx={BORDER_RADIUS}
              ry={BORDER_RADIUS}
              className={classNames(styles.connectedBorder)}
            />
          </>
        )}
        {!hideDetails && (
          <>
            {/* Egress URL */}
            <SVGButton
              access={node.access}
              isInteractive
              isSelected={pipelineSelected}
              aria-label={`${groupName} egress url`}
              onClick={() => onClick('egress')}
              round="top"
              transform={`translate (${BUTTON_MARGIN}, ${BUTTON_MARGIN - 1})`}
              IconSVG={
                !copied ? (
                  <CopySVG x="11" y="12" />
                ) : (
                  <CheckmarkSVG x="11" y="12" className={styles.checkmarkSvg} />
                )
              }
            />
            {/* Egress to */}
            <g
              transform={`translate (${BUTTON_MARGIN}, ${
                BUTTON_HEIGHT + BUTTON_MARGIN - 1
              })`}
              id={`${groupName}_to`}
            >
              <use
                xlinkHref="#bottomRoundedButton"
                clipPath="url(#bottomRoundedButtonOnly)"
                fill="#FAFAFA"
              />
              {node.egressType && (
                <text
                  style={labelTextStyle}
                  className={styles.subLabel}
                  x="9"
                  y="20"
                >
                  {`Egresses to ${node.egressType}`}
                </text>
              )}
            </g>
            <g
              transform={`translate (${FOOTER_LEFT_MARGIN}, ${
                CONNECTED_EGRESS_NODE_HEIGHT - FOOTER_HEIGHT + FOOTER_HEIGHT / 4
              })`}
            >
              <g transform="scale(0.75)">
                <EgressSVG fill="#666" />
              </g>
              <text style={labelTextStyle} x="25" y="12">
                Egress
              </text>
            </g>
          </>
        )}
      </g>
    );
  }

  if (node.type === NodeType.CROSS_PROJECT_REPO) {
    return (
      <g
        aria-label={`${groupName} cross project repo`}
        id={groupName}
        transform={`translate (${node.x}, ${node.y})`}
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
      >
        {/* Container rectangle */}
        <rect
          width={NODE_WIDTH}
          height={CONNECTED_EGRESS_NODE_HEIGHT}
          rx={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
          ry={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
          className={classNames(styles.node, {
            [styles.simplifiedInput]: showSimple,
            [styles.noShadow]: showSimple,
          })}
        />

        {!showSimple && (
          <>
            {/* Gradient - mask defined in DAGView.tsx */}
            <use
              xlinkHref="#connectedEgressMask"
              className={classNames(styles.greenGradient, {
                [styles.hideDetails]: hideDetails,
              })}
            />

            <rect
              width={NODE_WIDTH + BORDER_WIDTH}
              height={CONNECTED_EGRESS_NODE_HEIGHT + BORDER_WIDTH}
              transform="translate (-1, -1)"
              rx={BORDER_RADIUS}
              ry={BORDER_RADIUS}
              className={classNames(styles.connectedBorder)}
            />
          </>
        )}
        {!hideDetails && (
          <>
            {/* Connected Project */}
            <SVGButton
              access={node.access}
              isInteractive
              isSelected={pipelineSelected}
              aria-label={`${groupName} connected project`}
              onClick={() => onClick('connected_project')}
              round="top"
              transform={`translate (${BUTTON_MARGIN}, ${BUTTON_MARGIN - 1})`}
              IconSVG={<LinkSVG x="11" y="12" />}
              textClassName="nodeLabelProject"
            />

            {/* Connected Repo */}
            <SVGButton
              access={node.access}
              isInteractive
              isSelected={repoSelected}
              aria-label={`${groupName} connected repo`}
              id="connectedRepoGroup"
              onClick={() => onClick('connected_repo')}
              transform={`translate (${BUTTON_MARGIN}, ${
                BUTTON_HEIGHT + BUTTON_MARGIN - 1
              })`}
              round="bottom"
              IconSVG={<RepoSVG x="11" y="11" />}
            />
            <g
              transform={`translate (${FOOTER_LEFT_MARGIN}, ${
                CONNECTED_EGRESS_NODE_HEIGHT - FOOTER_HEIGHT + FOOTER_HEIGHT / 4
              })`}
            >
              <g transform="scale(0.75)">
                <PachProjectSVG fill="#666" />
              </g>
              <text style={labelTextStyle} x="25" y="12">
                Connected Project
              </text>
            </g>
          </>
        )}
      </g>
    );
  }

  return (
    <g
      id={groupName}
      transform={`translate (${node.x}, ${node.y})`}
      onMouseOver={onMouseOver}
      onMouseOut={onMouseOut}
    >
      <rect
        width={NODE_WIDTH}
        height={NODE_HEIGHT}
        className={classNames(styles.node, {
          [styles.detailedError]:
            !showSimple && node.nodeState === NodeState.ERROR,
          [styles.pipelineSimplifiedBox]:
            showSimple && node.jobNodeState !== NodeState.ERROR,
          [styles.pipelineSimplifiedBoxError]:
            showSimple && node.jobNodeState === NodeState.ERROR,
          [styles.noShadow]: showSimple,
        })}
        rx={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
        ry={showSimple ? SIMPLE_BORDER_RADIUS : BORDER_RADIUS}
      />
      {hideDetails && !showSimple && node.jobNodeState === NodeState.ERROR && (
        <g
          transform={` translate (${NODE_WIDTH / 2 - 38}, ${
            (NODE_HEIGHT - FOOTER_HEIGHT) / 2 - 17
          })
          scale (1.75)`}
        >
          <NodeStateIcon state={node.jobNodeState} x={0} y={0} />
          <JobsSVG x={25} y={0} />
        </g>
      )}

      {!showSimple && (
        <>
          {node.type !== NodeType.PIPELINE ? (
            <use
              xlinkHref="#pipelineMask"
              className={classNames(styles.purpleGradient, {
                [styles.hideDetails]: hideDetails,
              })}
            />
          ) : (
            <rect
              width={NODE_WIDTH - BUTTON_MARGIN * 2}
              height={NODE_HEIGHT - (FOOTER_HEIGHT + BUTTON_MARGIN * 2)}
              className={classNames(styles.pipelineInternalBorder, {
                [styles.simplified]: showSimple,
              })}
              rx={BORDER_RADIUS}
              ry={BORDER_RADIUS}
              transform={`translate (${BUTTON_MARGIN}, ${BUTTON_MARGIN})`}
            />
          )}
        </>
      )}

      {!hideDetails && (
        <>
          {/* buttons background */}
          <rect
            width={BUTTON_WIDTH}
            height={NODE_HEIGHT - (FOOTER_HEIGHT + BUTTON_MARGIN * 2)}
            className={styles.pipelineButtonsBackground}
            rx={BORDER_RADIUS}
            ry={BORDER_RADIUS}
            x={BUTTON_MARGIN}
            y={BUTTON_MARGIN}
          />

          {/* pipeline button */}
          <SVGButton
            access={node.access}
            isInteractive
            isSelected={pipelineSelected}
            onClick={() => onClick('pipeline')}
            aria-label={`${groupName} pipeline`}
            id="pipelineButtonGroup"
            round="top"
            transform={`translate (${BUTTON_MARGIN}, ${BUTTON_MARGIN - 1})`}
            IconSVG={<PipelineSVG x="11" y="12" />}
          />

          {node.hasCronInput && (
            <g
              transform={`
              translate (${NODE_WIDTH - 27}, 12)
              scale(0.75)`}
              className={styles.tooltipWrapper}
              id={`${groupName}_cronIcon`}
            >
              <PachCronSVG className={styles.pipelineCronIcon} />
              <rect width="20" height="20" fill="transparent">
                <title>CRON</title>
              </rect>
            </g>
          )}

          {/* repo button */}
          <SVGButton
            access={node.access}
            isInteractive
            isSelected={repoSelected}
            onClick={() => onClick('repo')}
            aria-label={`${groupName} repo`}
            id="repoButtonGroup"
            transform={`translate (${BUTTON_MARGIN}, ${
              BUTTON_MARGIN + BUTTON_HEIGHT
            })`}
            round="none"
            text="Output"
            IconSVG={<RepoSVG x="11" y="11" />}
          />

          {/* pipeline status button */}
          {node.state && (
            <g
              role="button"
              aria-label={`${groupName} status`}
              id="pipelineStatusGroup"
              data-testid={`Node__state-${node.nodeState}`}
              transform={`translate (${
                BUTTON_MARGIN * 2 + STATUS_BUTTON_WIDTH
              }, ${BUTTON_MARGIN * 1.5 + BUTTON_HEIGHT * 2})`}
              onClick={() => onClick('status')}
              className={statusClasses}
            >
              <rect
                width={STATUS_BUTTON_WIDTH}
                height={STATUS_BUTTON_HEIGHT}
                rx={BORDER_RADIUS}
                ry={BORDER_RADIUS}
                className={styles.statusRect}
              />
              <text style={labelTextStyle} x={BUTTON_MARGIN * 2 + 2} y={18}>
                Pipeline
              </text>

              <text
                x={BUTTON_MARGIN * 2 + 20}
                y={40}
                className={statusTextClasses(node.nodeState)}
              >
                {node.state && readablePipelineState(node.state)}
              </text>
              <g transform="scale(0.6)">
                <NodeStateIcon
                  state={node.nodeState}
                  x={(BUTTON_MARGIN * 2 + 2) / 0.6}
                  y={48}
                />
              </g>
            </g>
          )}

          {/* job status button */}
          {node.jobNodeState && node.jobNodeState !== NodeState.IDLE && (
            <g
              role="button"
              aria-label={`${groupName} logs`}
              id="jobStatusGroup"
              data-testid={`Node__state-${node.jobNodeState}`}
              transform={`translate (${BUTTON_MARGIN}, ${
                BUTTON_MARGIN * 1.5 + BUTTON_HEIGHT * 2
              })`}
              onClick={() => onClick('logs')}
              className={statusClasses}
            >
              <rect
                width={STATUS_BUTTON_WIDTH}
                height={STATUS_BUTTON_HEIGHT}
                rx={BORDER_RADIUS}
                ry={BORDER_RADIUS}
                className={styles.statusRect}
              />
              <text
                style={labelTextStyle}
                className={styles.subLabel}
                x={BUTTON_MARGIN * 2 + 2}
                y={18}
              >
                Subjob
              </text>

              <text
                x={BUTTON_MARGIN * 2 + 20}
                y={40}
                className={statusTextClasses(node.jobNodeState)}
              >
                {node.jobState && readableJobState(node.jobState)}
              </text>
              <g transform="scale(0.6)">
                <NodeStateIcon
                  state={node.jobNodeState}
                  x={(BUTTON_MARGIN * 2 + 2) / 0.6}
                  y={48}
                />
              </g>
            </g>
          )}

          <g
            transform={`translate (${FOOTER_LEFT_MARGIN}, ${
              NODE_HEIGHT - FOOTER_HEIGHT + FOOTER_HEIGHT / 4
            })`}
          >
            {parallelismInfo(`${groupName}_parallelism`, node.parallelism)}
            {node.type !== NodeType.PIPELINE && (
              <>
                <g transform="scale(0.75)">
                  <PipelineFooterIcon type={node.type} />
                </g>
                <text style={labelTextStyle} x="25" y="12">
                  {pipelineFooterText(node.type)}
                </text>
              </>
            )}
          </g>
        </>
      )}
    </g>
  );
};

export default Node;
