import React from 'react';

import styles from './Mark.module.css';

type MarkProps = {
  children?: React.ReactNode;
  type?: 'pipeline' | 'repo';
};

const Mark: React.FC<MarkProps> = ({children}) => (
  <span className={styles.base}>{children}</span>
);

export default Mark;
