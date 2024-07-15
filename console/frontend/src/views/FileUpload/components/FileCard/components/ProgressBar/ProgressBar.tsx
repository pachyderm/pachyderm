import React, {ProgressHTMLAttributes} from 'react';

import {Group} from '@pachyderm/components';

import styles from './ProgressBar.module.css';

const ProgressBar: React.FC<ProgressHTMLAttributes<HTMLProgressElement>> = ({
  value,
  ...rest
}) => {
  return (
    <Group align="center" spacing={8}>
      <progress className={styles.base} {...rest} value={value} />
      <span className={styles.percentage}>{value}%</span>
    </Group>
  );
};

export default ProgressBar;
