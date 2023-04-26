import React from 'react';

import styles from './TableViewPaginationWrapper.module.css';
const TableViewPaginationWrapper: React.FC = ({children}) => {
  return <div className={styles.base}>{children}</div>;
};
export default TableViewPaginationWrapper;
