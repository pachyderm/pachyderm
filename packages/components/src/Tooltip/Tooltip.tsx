import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useCallback} from 'react';
import OverlayTrigger, {
  OverlayTriggerProps,
} from 'react-bootstrap/OverlayTrigger';
import BootstrapTooltip from 'react-bootstrap/Tooltip';

import {ExitSVG} from '../Svg';

import styles from './Tooltip.module.css';

interface TooltipProps
  extends Omit<OverlayTriggerProps, 'overlay' | 'ref' | 'key'> {
  size?: 'large' | 'extraLarge';
  disabled?: boolean;
  tooltipText: string | React.ReactNode;
  tooltipKey: string;
  delay?: {show: number; hide: number};
  showCloseButton?: boolean;
  className?: string;
}

const Tooltip: React.FC<TooltipProps> = ({
  size = '',
  disabled = false,
  tooltipKey,
  tooltipText,
  children,
  delay = {show: 100, hide: 200},
  onToggle,
  showCloseButton = false,
  className,
  ...rest
}) => {
  const renderTooltip = useCallback(
    (props) => {
      if (disabled) {
        return <div aria-hidden={true} />;
      }

      return (
        <BootstrapTooltip
          id={`tooltip-${tooltipKey}`}
          bs-prefix={'tooltip'}
          className={classNames(styles.base, styles[size], className, {
            [styles.hasClose]: showCloseButton,
          })}
          data-testid={`tooltip-${tooltipKey}`}
          delay={delay}
          {...props}
        >
          {tooltipText}

          {showCloseButton && (
            <button
              aria-label="Close"
              data-testid={`tooltip-${tooltipKey}-close`}
              onClick={onToggle || noop}
              className={styles.close}
            >
              <ExitSVG className={styles.icon} />
            </button>
          )}
        </BootstrapTooltip>
      );
    },
    [
      className,
      delay,
      size,
      disabled,
      tooltipKey,
      tooltipText,
      onToggle,
      showCloseButton,
    ],
  );

  return (
    <OverlayTrigger
      key={tooltipKey}
      overlay={renderTooltip}
      onToggle={onToggle}
      {...rest}
    >
      {children}
    </OverlayTrigger>
  );
};

export default Tooltip;
