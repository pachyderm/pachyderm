import classnames from 'classnames';
import React from 'react';

import styles from './Tooltip.module.css';

interface TooltipProps {
  children?: React.ReactNode;
  show: boolean;
  x?: number;
  y?: number;
  width: number;
  height: number;
  hasArrow?: boolean;
  noCaps?: boolean;
}

const Tooltip: React.FC<TooltipProps> = ({
  show,
  x = 0,
  y = 0,
  width,
  height,
  hasArrow,
  noCaps,
  children,
}) => {
  return (
    <foreignObject
      className={styles.base}
      width={width}
      height={height}
      y={y}
      x={x}
    >
      <span>
        <p
          className={classnames(styles.text, {
            [styles.show]: show,
            [styles.noCaps]: noCaps,
            [styles.hasArrow]: hasArrow,
          })}
          style={{minHeight: `${height - 9}px`}}
        >
          {children}
        </p>
      </span>
    </foreignObject>
  );
};

export default Tooltip;
