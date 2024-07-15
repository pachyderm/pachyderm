import classNames from 'classnames';
import React from 'react';

import styles from './TableViewWrapper.module.css';

type TableViewWrapperProps = {
  children?: React.ReactNode;
  hasPager?: boolean;
};

const TableViewWrapper: React.FC<TableViewWrapperProps> = ({
  children,
  hasPager,
}) => {
  return (
    <div className={classNames(styles.base, {[styles.hasPager]: hasPager})}>
      {children}
    </div>
  );
};

export default TableViewWrapper;
