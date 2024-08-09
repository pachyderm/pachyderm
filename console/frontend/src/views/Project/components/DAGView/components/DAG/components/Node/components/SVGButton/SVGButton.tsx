import classNames from 'classnames';
import React, {SVGProps} from 'react';

import {
  BORDER_RADIUS,
  BUTTON_HEIGHT,
  BUTTON_WIDTH,
} from '@dash-frontend/views/Project/constants/nodeSizes';
import {LockSVG} from '@pachyderm/components';

import styles from './SVGButton.module.css';

interface SVGButtonProps extends SVGProps<SVGGElement> {
  access: boolean;
  isInteractive: boolean;
  isSelected: boolean;
  round?: 'top' | 'bottom' | 'none';
  text?: string;
  title?: string;
  textClassName?: string;
  IconSVG?: JSX.Element;
}

const SVGButton: React.FC<SVGButtonProps> = ({
  access,
  isInteractive,
  isSelected,
  round,
  text,
  title,
  IconSVG,
  textClassName = 'nodeLabel',
  ...rest
}) => {
  const classes = classNames(styles.buttonGroup, {
    [styles.interactive]: isInteractive,
    [styles.selected]: isSelected,
  });

  const textElementProps = {
    fontSize: '14px',
    fontWeight: '600',
    textAnchor: 'start',
    dominantBaseline: 'middle',
    className: textClassName,
  };

  let container = (
    <rect
      className={styles.roundedButton}
      width={BUTTON_WIDTH}
      height={BUTTON_HEIGHT}
      rx={round === 'none' ? 0 : BORDER_RADIUS}
      ry={round === 'none' ? 0 : BORDER_RADIUS}
    />
  );

  // These paths are defiend in DAGView.tsx
  if (round) {
    if (round === 'top') {
      container = (
        <use
          transform="translate (0,-1)"
          xlinkHref="#topRoundedButton"
          clipPath="url(#topRoundedButtonOnly)"
          className={styles.roundedButton}
        />
      );
    } else if (round === 'bottom') {
      container = (
        <use
          transform="translate (0,-1)"
          xlinkHref="#bottomRoundedButton"
          clipPath="url(#bottomRoundedButtonOnly)"
          className={styles.roundedButton}
        />
      );
    }
  }

  return (
    <g role="button" className={classes} {...rest}>
      {container}
      <text {...textElementProps} x="34" y="17">
        {text}
        <title>{title}</title>
      </text>
      <g transform="scale(0.75)">
        {access ? (
          <>{IconSVG}</>
        ) : (
          <LockSVG color="var(--disabled-tertiary)" x="15" y="11" />
        )}
      </g>
    </g>
  );
};

export default SVGButton;
