import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  useCallback,
} from 'react';
import {useFormContext} from 'react-hook-form';

import useRHFInputProps from '@pachyderm/components/hooks/useRHFInputProps';

import styles from './Chip.module.css';
export interface ChipInputProps extends InputHTMLAttributes<HTMLInputElement> {
  name: string;
}

export interface ChipProps<T = unknown>
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onClick'> {
  selected?: boolean;
  onClick?: (value?: T) => void;
  onClickValue?: T;
}

export const Chip = <T,>({
  className,
  children,
  selected,
  onClick = noop,
  onClickValue,
  disabled,
  ...rest
}: React.PropsWithChildren<ChipProps<T>>) => {
  const classes = classNames(styles.base, className, {
    [styles.selected]: selected,
  });

  const onClickCallback = useCallback(
    (value?: T) => {
      onClick(value);
    },
    [onClick],
  );

  return (
    <button
      disabled={disabled}
      className={classes}
      aria-pressed={selected}
      onClick={() => onClickCallback(onClickValue)}
      {...rest}
    >
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
  disabled,
  ...rest
}) => {
  const {register, watch} = useFormContext();
  const value = watch(name);
  const classes = classNames(styles.inputBase, className, {
    [styles.selected]: value,
    [styles.disabled]: disabled,
  });

  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name),
  });

  return (
    <label className={classes}>
      <input
        disabled={disabled}
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
