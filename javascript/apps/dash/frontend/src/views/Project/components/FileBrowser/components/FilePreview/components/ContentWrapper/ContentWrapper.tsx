import React from 'react';

import styles from './ContentWrapper.module.css';

const ContentWrapper: React.FC = ({children}) => {
  return <div className={styles.base}>{children}</div>;
};

export default ContentWrapper;
