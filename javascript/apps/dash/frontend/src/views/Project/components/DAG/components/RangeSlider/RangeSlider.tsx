import React from 'react';

import styles from './RangeSlider.module.css';

type DagProps = {
  min: string;
  max: string;
  handleChange: React.ChangeEventHandler<HTMLInputElement> | undefined;
  value: number;
};

const DAG: React.FC<DagProps> = ({min, max, handleChange, value}) => {
  return (
    <input
      type="range"
      min={min}
      max={max}
      onChange={handleChange}
      value={value}
      className={styles.base}
    />
  );
};

export default DAG;
