import React from 'react';

import styles from './TableViewPaginationWrapper.module.css';

const TableViewPaginationWrapper = ({
  children,
}: {
  children?: React.ReactNode;
}) => {
  return <div className={styles.base}>{children}</div>;
};

export default TableViewPaginationWrapper;
