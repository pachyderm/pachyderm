import React from 'react';

import styles from './Title.module.css';

const Title = ({children}: {children?: React.ReactNode}) => (
  <h5 className={styles.base} data-testid="Title__name">
    {children}
  </h5>
);

export default Title;
