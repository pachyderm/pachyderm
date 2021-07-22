import React from 'react';

import styles from './Title.module.css';

const Title: React.FC = ({children}) => (
  <h2 className={styles.base} data-testid={'Title__name'}>
    {children}
  </h2>
);

export default Title;
