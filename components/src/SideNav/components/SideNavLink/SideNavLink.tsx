import classnames from 'classnames';
import React from 'react';
import {NavLink, NavLinkProps} from 'react-router-dom';

import {Group} from '../../../Group';
import {Icon} from '../../../Icon';
import {Tooltip} from '../../../Tooltip';
import useSideNav from '../../hooks/useSideNav';

import styles from './SideNavLink.module.css';

export interface SideNavLinkProps extends NavLinkProps {
  ['data-testid']?: string;
  IconSVG: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  disabled?: boolean;
  showIconWhenExpanded?: boolean;
  styleMode?: 'light' | 'dark';
  tooltipContent: string;
}

const SideNavLink: React.FC<SideNavLinkProps> = ({
  'data-testid': dataTestId,
  to,
  children,
  IconSVG,
  disabled,
  styleMode = 'dark',
  showIconWhenExpanded = false,
  tooltipContent,
  ...rest
}) => {
  const {minimized} = useSideNav();

  return (
    <Tooltip
      tooltipKey={tooltipContent}
      placement="right"
      disabled={!minimized}
      tooltipText={tooltipContent}
    >
      <NavLink
        to={to}
        className={classnames(styles.base, {
          [styles.disabled]: disabled,
          [styles[styleMode]]: true,
          [styles.minimized]: minimized,
          [styles.showIconWhenExpanded]: showIconWhenExpanded && !minimized,
        })}
        data-testid={dataTestId}
        {...rest}
      >
        <Group spacing={8}>
          {(showIconWhenExpanded || minimized) && (
            <Icon className={styles.icon}>
              <IconSVG />
            </Icon>
          )}
          {!minimized && children}
        </Group>
      </NavLink>
    </Tooltip>
  );
};

export default SideNavLink;
