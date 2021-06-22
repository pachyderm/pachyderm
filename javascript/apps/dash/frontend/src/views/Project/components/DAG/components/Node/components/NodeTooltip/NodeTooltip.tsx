import classnames from 'classnames';
import React from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Node, NodeType} from '@graphqlTypes';
import {deriveNameFromNodeNameAndType} from 'lib/deriveRepoNameFromNode';

import styles from './NodeTooltip.module.css';

const NODE_TOOLTIP_WIDTH = 250;
const NODE_TOOLTIP_HEIGHT = 80;
const NODE_TOOLTIP_Y_OFFSET = -92;
const NODE_TOOLTIP_X_OFFSET = -(NODE_TOOLTIP_WIDTH / 4);
const EGRESS_NODE_TOOLTIP_Y_OFFSET = NODE_TOOLTIP_Y_OFFSET + 22;
const EGRESS_NODE_TOOLTIP_X_OFFSET = NODE_TOOLTIP_X_OFFSET - 23;

interface TooltipContentProps {
  name: Node['name'];
  type: Node['type'];
  state: Node['state'];
  access: Node['access'];
}

const TooltipContent: React.FC<TooltipContentProps> = ({
  name,
  type,
  state,
  access,
}) => {
  if (!access) {
    // NOTE: in the future, there may be auth on nodes
    // other than repos.
    return (
      <>
        In order to view this resource, you must have the &quot;repoReader&quot;
        role.
      </>
    );
  }

  if (type === NodeType.EGRESS) {
    return <>Click to copy S3 address: {name}</>;
  }

  return (
    <>
      {' '}
      {deriveNameFromNodeNameAndType(name, type)}
      <br />
      <br />
      {type === NodeType.PIPELINE
        ? `${type.toLowerCase()} status: ${readablePipelineState(state || '')}`
        : ''}
    </>
  );
};

interface NodeTooltipProps {
  node: Node;
  show: boolean;
}

const NodeTooltip: React.FC<NodeTooltipProps> = ({
  node: {type, name, state, access},
  show,
}) => {
  const isEgress = type === NodeType.EGRESS;

  return (
    <foreignObject
      className={styles.base}
      width={NODE_TOOLTIP_WIDTH}
      height={NODE_TOOLTIP_HEIGHT}
      y={isEgress ? EGRESS_NODE_TOOLTIP_Y_OFFSET : NODE_TOOLTIP_Y_OFFSET}
      x={isEgress ? EGRESS_NODE_TOOLTIP_X_OFFSET : NODE_TOOLTIP_X_OFFSET}
    >
      <span>
        <p
          className={classnames(styles.text, {
            [styles.noCaps]: isEgress || !access,
            [styles.show]: show,
          })}
          style={{minHeight: `${NODE_TOOLTIP_HEIGHT - 9}px`}}
        >
          <TooltipContent
            name={name}
            type={type}
            state={state}
            access={access}
          />
        </p>
      </span>
    </foreignObject>
  );
};

export default NodeTooltip;
