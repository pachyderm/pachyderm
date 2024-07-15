import React from 'react';

import styles from './ContentWrapper.module.css';

const ContentWrapper = ({children}: {children?: React.ReactNode}) => {
  return <div className={styles.base}>{children}</div>;
};

export default ContentWrapper;
