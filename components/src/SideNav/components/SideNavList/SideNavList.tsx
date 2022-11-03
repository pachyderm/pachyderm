import classNames from 'classnames';
import React from 'react';

import styles from './SideNavList.module.css';

type SideNavListProps = {
  noPadding?: boolean;
};

const SideNavList: React.FC<SideNavListProps> = ({
  noPadding = false,
  children,
}) => {
  return (
    <ul className={classNames(styles.base, {[styles.noPadding]: noPadding})}>
      {children}
    </ul>
  );
};

export default SideNavList;
