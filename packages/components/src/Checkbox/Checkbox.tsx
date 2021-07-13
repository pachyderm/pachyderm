import classNames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {
  useFormContext,
  RegisterOptions,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import useRHFInputProps from 'hooks/useRHFInputProps';

import {CheckboxCheckedSVG, CheckboxSVG} from '../Svg';

import styles from './Checkbox.module.css';

export interface CheckboxProps extends InputHTMLAttributes<HTMLInputElement> {
  name: FieldPath<FieldValues>;
  label?: React.ReactNode;
  small?: boolean;
  validationOptions?: RegisterOptions;
}

export interface PureCheckboxProps
  extends InputHTMLAttributes<HTMLInputElement> {
  selected: boolean;
  label?: React.ReactNode;
  small?: boolean;
  disabled?: boolean;
}

export const Checkbox: React.FC<CheckboxProps> = ({
  label,
  name,
  className,
  small = false,
  validationOptions = {},
  onChange,
  onBlur,
  disabled = false,
  ...rest
}) => {
  const {register, watch} = useFormContext();

  const value = watch(name);

  const classes = classNames(styles.base, className, {
    [styles.small]: small,
    [styles.disabled]: disabled,
  });

  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name, validationOptions),
  });

  return (
    <label className={classes}>
      <div className={styles.checkboxContainer}>
        <input
          type="checkbox"
          className={styles.input}
          onChange={handleChange}
          onBlur={handleBlur}
          disabled={disabled}
          {...rest}
          {...inputProps}
        />

        {!value && (
          <CheckboxSVG
            className={styles.checkbox}
            aria-hidden
            focusable={false}
          />
        )}

        {value && (
          <CheckboxCheckedSVG
            className={styles.checked}
            aria-hidden
            focusable={false}
          />
        )}
      </div>

      <span className={styles.label}>{label}</span>
    </label>
  );
};

export const PureCheckbox: React.FC<PureCheckboxProps> = ({
  selected,
  label,
  small = false,
  className,
  disabled = false,
  ...rest
}) => {
  const classes = classNames(styles.base, className, {
    [styles.small]: small,
    [styles.disabled]: disabled,
  });

  return (
    <label className={classes}>
      <div className={styles.checkboxContainer}>
        <input
          type="checkbox"
          className={styles.input}
          disabled={disabled}
          {...rest}
        />

        {!selected && (
          <CheckboxSVG
            className={styles.checkbox}
            aria-hidden
            focusable={false}
          />
        )}

        {selected && (
          <CheckboxCheckedSVG
            className={styles.checked}
            aria-hidden
            focusable={false}
          />
        )}
      </div>

      <span className={styles.label}>{label}</span>
    </label>
  );
};
