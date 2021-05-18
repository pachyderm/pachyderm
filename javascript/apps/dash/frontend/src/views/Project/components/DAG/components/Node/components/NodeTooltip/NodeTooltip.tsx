import classnames from 'classnames';
import React from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {NodeType, PipelineState} from '@graphqlTypes';

import {deriveNameFromNodeNameAndType} from '../../../../utils/deriveRepoNameFromNode';

import styles from './NodeTooltip.module.css';

const NODE_TOOLTIP_WIDTH = 250;
const NODE_TOOLTIP_HEIGHT = 80;
const NODE_TOOLTIP_Y_OFFSET = -92;
const NODE_TOOLTIP_X_OFFSET = -(NODE_TOOLTIP_WIDTH / 4);
const EGRESS_NODE_TOOLTIP_Y_OFFSET = NODE_TOOLTIP_Y_OFFSET + 22;
const EGRESS_NODE_TOOLTIP_X_OFFSET = NODE_TOOLTIP_X_OFFSET - 23;

interface NodeTooltipProps {
  nodeType: NodeType;
  nodeName: string;
  nodeState?: PipelineState | null;
  textClassName?: string;
}

const NodeTooltip: React.FC<NodeTooltipProps> = ({
  nodeType,
  nodeName,
  nodeState = '',
  textClassName = '',
}) => {
  const isEgress = nodeType === NodeType.EGRESS;

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
          className={classnames(styles.text, textClassName, {
            [styles.egress]: isEgress,
          })}
          style={{minHeight: `${NODE_TOOLTIP_HEIGHT - 9}px`}}
        >
          {isEgress ? (
            <>Click to copy S3 address: {nodeName}</>
          ) : (
            <>
              {' '}
              {deriveNameFromNodeNameAndType(nodeName, nodeType)}
              <br />
              <br />
              {nodeType === NodeType.PIPELINE
                ? `${nodeType.toLowerCase()} status: ${readablePipelineState(
                    nodeState || '',
                  )}`
                : ''}
            </>
          )}
        </p>
      </span>
    </foreignObject>
  );
};

export default NodeTooltip;
