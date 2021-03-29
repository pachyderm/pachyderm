import classNames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {useFormContext} from 'react-hook-form';

import styles from './Chip.module.css';

export interface ChipInputProps extends InputHTMLAttributes<HTMLInputElement> {
  name: string;
  label?: string;
  'data-testid'?: string;
}

export type ChipButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  'data-testid'?: string;
};

export const Chip: React.FC<ChipButtonProps> = ({
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
  label,
  className,
  ...rest
}) => {
  const {register, watch} = useFormContext();
  const value = watch(name);
  const classes = classNames(styles.base, className, {
    [styles.selected]: value,
  });

  return (
    <label className={classes}>
      <input
        ref={register()}
        type="checkbox"
        className={styles.input}
        name={name}
        {...rest}
      />
      {label ? label : name}
    </label>
  );
};

export const ChipGroup: React.FC = ({children}) => {
  return <div className={styles.group}>{children}</div>;
};
