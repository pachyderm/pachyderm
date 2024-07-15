import React from 'react';

import styles from './RangeSlider.module.css';

interface RangeSliderProps
  extends Omit<
    React.InputHTMLAttributes<HTMLInputElement>,
    'type' | 'className' | 'min' | 'max' | 'number' | 'value'
  > {
  min: string;
  max: string;
  value: number;
}

const RangeSlider: React.FC<RangeSliderProps> = ({
  min,
  max,
  onChange,
  value,
  ...rest
}) => {
  return (
    <input
      type="range"
      min={min}
      max={max}
      onChange={onChange}
      value={value}
      className={styles.base}
      {...rest}
    />
  );
};

export default RangeSlider;
