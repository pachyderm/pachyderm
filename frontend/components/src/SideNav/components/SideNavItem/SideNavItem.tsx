import classnames from 'classnames';
import React from 'react';
import {NavLink, NavLinkProps} from 'react-router-dom';

import {ButtonLink} from '../../../ButtonLink';
import {Group} from '../../../Group';
import {Icon} from '../../../Icon';
import {Tooltip} from '../../../Tooltip';
import useSideNav from '../../hooks/useSideNav';

import styles from './SideNavItem.module.css';

interface SideNavItemProps extends Omit<NavLinkProps, 'to'> {
  ['data-testid']?: string;
  IconSVG: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  disabled?: boolean;
  showIconWhenExpanded?: boolean;
  tooltipContent: string;
  onClick?: () => void;
  to?: string;
}

const SideNavItem: React.FC<SideNavItemProps> = ({
  'data-testid': dataTestId,
  to,
  children,
  IconSVG,
  disabled,
  showIconWhenExpanded = false,
  tooltipContent,
  onClick,
}) => {
  const {minimized} = useSideNav();
  const itemClassNames = classnames(styles.item, {
    [styles.disabled]: disabled,
    [styles.minimized]: minimized,
    [styles.showIconWhenExpanded]: showIconWhenExpanded && !minimized,
  });

  const itemContent = (
    <Group spacing={8}>
      {(showIconWhenExpanded || minimized) && (
        <Icon className={styles.icon} small>
          <IconSVG />
        </Icon>
      )}
      {!minimized && children}
    </Group>
  );

  const tooltipWrapper = (children: JSX.Element) => (
    <Tooltip
      allowedPlacements={['right']}
      noWrap
      disabled={!minimized}
      tooltipText={tooltipContent}
    >
      {children}
    </Tooltip>
  );

  return (
    <li className={styles.base}>
      {onClick &&
        tooltipWrapper(
          <ButtonLink
            onClick={onClick}
            className={itemClassNames}
            data-testid={dataTestId}
          >
            {itemContent}
          </ButtonLink>,
        )}
      {to &&
        tooltipWrapper(
          <NavLink to={to} className={itemClassNames} data-testid={dataTestId}>
            {itemContent}
          </NavLink>,
        )}
    </li>
  );
};

export default SideNavItem;
