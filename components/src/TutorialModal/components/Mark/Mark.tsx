import React from 'react';

import styles from './Mark.module.css';

type MarkProps = {
  type?: 'pipeline' | 'repo';
};

const Mark: React.FC<MarkProps> = ({type, children}) => (
  <span className={styles.base}>{children}</span>
);

export default Mark;
