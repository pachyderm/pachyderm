import classNames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {useFormContext} from 'react-hook-form';

import useRHFInputProps from 'hooks/useRHFInputProps';

import styles from './Chip.module.css';

export interface ChipInputProps extends InputHTMLAttributes<HTMLInputElement> {
  name: string;
}

export const Chip: React.FC<React.ButtonHTMLAttributes<HTMLButtonElement>> = ({
  className,
  children,
  ...rest
}) => {
  const classes = classNames(styles.base, className);
  return (
    <button className={classes} {...rest}>
      {children}
    </button>
  );
};

export const ChipInput: React.FC<ChipInputProps> = ({
  name,
  className,
  children,
  onChange,
  onBlur,
  ...rest
}) => {
  const {register, watch} = useFormContext();
  const value = watch(name);
  const classes = classNames(styles.base, className, {
    [styles.selected]: value,
  });
  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name),
  });

  return (
    <label className={classes}>
      <input
        type="checkbox"
        className={styles.input}
        onChange={handleChange}
        onBlur={handleBlur}
        {...rest}
        {...inputProps}
      />
      {children}
    </label>
  );
};

export const ChipGroup: React.FC = ({children}) => {
  return <div className={styles.group}>{children}</div>;
};
