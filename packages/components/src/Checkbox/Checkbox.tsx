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

export const Checkbox: React.FC<CheckboxProps> = ({
  label,
  name,
  className,
  small = false,
  validationOptions = {},
  onChange,
  onBlur,
  ...rest
}) => {
  const {register, watch} = useFormContext();

  const value = watch(name);

  const classes = classNames(styles.base, className, {
    [styles.small]: small,
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
