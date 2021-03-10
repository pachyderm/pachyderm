import React from 'react';

import styles from './CardHeader.module.css';

const CardHeader: React.FC = ({children}) => {
  return <div className={styles.base}>{children}</div>;
};

export default CardHeader;
