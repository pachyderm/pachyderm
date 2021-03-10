import classNames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {useFormContext, RegisterOptions} from 'react-hook-form';

import {CheckboxCheckedSVG, CheckboxSVG} from '../Svg';

import styles from './Checkbox.module.css';

export interface CheckboxProps extends InputHTMLAttributes<HTMLInputElement> {
  name: string;
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
  ...rest
}) => {
  const {register, watch} = useFormContext();

  const value = watch(name);

  const classes = classNames(styles.base, className, {
    [styles.small]: small,
  });

  return (
    <label className={classes}>
      <div className={styles.checkboxContainer}>
        <input
          ref={register(validationOptions)}
          type="checkbox"
          className={styles.input}
          name={name}
          {...rest}
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
