import React from 'react';

import styles from './Title.module.css';

const Title: React.FC = ({children}) => (
  <h5 className={styles.base} data-testid="Title__name">
    {children}
  </h5>
);

export default Title;
