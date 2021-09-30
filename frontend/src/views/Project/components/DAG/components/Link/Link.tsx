import classNames from 'classnames';
import React from 'react';

import {DagDirection, Link as LinkType, Node} from '@dash-frontend/lib/types';

import Tooltip from '../Tooltip';

import useLink from './hooks/useLink';
import styles from './Link.module.css';

const NODE_TOOLTIP_WIDTH = 240;
const NODE_TOOLTIP_HEIGHT = 72;
const NODE_TOOLTIP_Y_OFFSET = 90;
const NODE_TOOLTIP_X_OFFSET = 0;

type LinkProps = {
  nodes: Node[];
  link: LinkType;
  isInteractive: boolean;
  dagDirection: DagDirection;
};

const Link: React.FC<LinkProps> = ({
  link,
  isInteractive,
  dagDirection,
  nodes,
}) => {
  const {
    d,
    isSelected,
    selectedNode,
    oppositeNode,
    transferring,
    shouldShowClickableLink,
    sourceSelected,
    onClick,
    fastForwardIconPoint,
    fastForwardIconOrientation,
    isHovering,
    setIsHovering,
    pathRef,
  } = useLink(link, isInteractive, dagDirection, nodes);

  const classes = classNames(
    styles.link,
    styles[`${link.sourceState?.toLowerCase()}Source`],
    styles[`${link.targetState?.toLowerCase()}Target`],
    {
      [styles.transferring]: transferring,
      [styles.selected]: isSelected,
      [styles.hovering]: isHovering,
    },
  );

  const iconPoint = fastForwardIconPoint();

  return (
    <g>
      <path ref={pathRef} className={classes} d={d} id={link.id} fill="none" />
      {shouldShowClickableLink && (
        <>
          <path
            className={styles.hoverLink}
            d={d}
            id={link.id}
            fill="none"
            onClick={onClick}
            onMouseEnter={() => setIsHovering(true)}
            onMouseLeave={() => setIsHovering(false)}
          />
          <Tooltip
            width={NODE_TOOLTIP_WIDTH}
            height={NODE_TOOLTIP_HEIGHT}
            y={(selectedNode?.y || 0) + NODE_TOOLTIP_Y_OFFSET}
            x={(selectedNode?.x || 0) + NODE_TOOLTIP_X_OFFSET}
            show={isHovering}
          >
            {oppositeNode
              ? `${sourceSelected ? 'Input from' : 'Output to'} ${
                  oppositeNode.name
                }`
              : ''}
          </Tooltip>
          {isHovering && (
            <>
              <rect
                className={styles.fastForwardIconWrapper}
                x={iconPoint.x - 12}
                y={iconPoint.y - 12}
                rx={2}
                width={24}
                height={24}
                onClick={onClick}
                onMouseEnter={() => setIsHovering(true)}
                onMouseLeave={() => setIsHovering(false)}
              />
              <image
                className={styles.fastForwardIcon}
                href="/dag_fast_forward.svg"
                x={iconPoint.x - 7}
                y={iconPoint.y - 6}
                transform={fastForwardIconOrientation(iconPoint)}
              />
            </>
          )}
        </>
      )}
      {transferring && (
        <circle r={6} className={styles.circle}>
          <animateMotion dur="0.8s" repeatCount="indefinite" path={d} />
        </circle>
      )}
    </g>
  );
};

export default Link;
