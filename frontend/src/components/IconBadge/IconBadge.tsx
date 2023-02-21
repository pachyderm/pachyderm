import classnames from 'classnames';
import React from 'react';
import {Link, LinkProps} from 'react-router-dom';

import {Icon, Tooltip} from '@pachyderm/components';

import styles from './IconBadge.module.css';

type IconBadgeProps = {
  color: 'red' | 'green' | 'black';
  'aria-label'?: string;
  IconSVG?: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  tooltip?: React.ReactNode;
  to?: LinkProps['to'];
};

const IconBadge: React.FC<IconBadgeProps> = ({
  children,
  color,
  IconSVG,
  tooltip = '',
  'aria-label': ariaLabel,
  to,
}) => {
  const IconBadgeWrapper: React.FC = ({children}) => {
    if (to) {
      return (
        <Link to={to} onClick={(e) => e.currentTarget.blur()}>
          {children}
        </Link>
      );
    } else {
      return <>{children}</>;
    }
  };

  return (
    <Tooltip
      disabled={!tooltip}
      tooltipKey={String(tooltip)}
      tooltipText={tooltip}
      placement="top"
    >
      <div
        className={classnames(styles.base, {
          [styles[color]]: true,
        })}
        aria-label={ariaLabel}
      >
        <IconBadgeWrapper>
          {IconSVG && (
            <Icon color={color} small className={styles.icon}>
              <IconSVG />
            </Icon>
          )}
          {children}
        </IconBadgeWrapper>
      </div>
    </Tooltip>
  );
};

export default IconBadge;
