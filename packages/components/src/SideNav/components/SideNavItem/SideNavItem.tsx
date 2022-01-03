import classnames from 'classnames';
import React from 'react';

import styles from './SideNavItem.module.css';

type Props = {
  secondary?: boolean;
};

const SideNavItem: React.FC<Props> = ({
  children,
  secondary = false,
  ...rest
}) => {
  return (
    <li
      className={classnames(styles.base, {[styles.secondary]: secondary})}
      {...rest}
    >
      {children}
    </li>
  );
};

export default SideNavItem;
