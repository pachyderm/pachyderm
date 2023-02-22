import React from 'react';

import styles from './TableViewWrapper.module.css';

const TableViewWrapper: React.FC = ({children}) => {
  return <div className={styles.base}>{children}</div>;
};

export default TableViewWrapper;
