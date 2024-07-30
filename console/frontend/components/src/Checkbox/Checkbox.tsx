import classNames from 'classnames';
import React, {InputHTMLAttributes, forwardRef} from 'react';
import {
  useFormContext,
  RegisterOptions,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import useRHFInputProps from '@pachyderm/components/hooks/useRHFInputProps';

import {Icon} from '../Icon';
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
}

export const Checkbox: React.FC<CheckboxProps> = ({
  name,
  validationOptions = {},
  onChange,
  onBlur,
  id,
  ...rest
}) => {
  const {register, watch} = useFormContext();
  const value = watch(name);
  let selected = value;

  if (Array.isArray(value)) {
    selected = value.includes(id);
  }

  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name, validationOptions),
  });

  return (
    <PureCheckbox
      onChange={handleChange}
      onBlur={handleBlur}
      selected={selected}
      {...rest}
      {...inputProps}
    />
  );
};

export const PureCheckbox = forwardRef<HTMLInputElement, PureCheckboxProps>(
  function PureCheckboxForwardRef(
    {
      className,
      selected,
      label,
      onChange,
      onBlur,
      small = false,
      disabled = false,
      ...rest
    },
    ref,
  ) {
    const classes = classNames(styles.base, className, {
      [styles.small]: small,
      [styles.disabled]: disabled,
    });

    return (
      <label className={classes} data-disabled={disabled}>
        <div className={styles.checkboxContainer}>
          <input
            type="checkbox"
            className={styles.input}
            disabled={disabled}
            onChange={onChange}
            onBlur={onBlur}
            ref={ref}
            {...rest}
          />
          <Icon small={small}>
            {selected ? (
              <CheckboxCheckedSVG aria-hidden focusable={false} />
            ) : (
              <CheckboxSVG aria-hidden focusable={false} />
            )}
          </Icon>
        </div>

        <span className={styles.label}>{label}</span>
      </label>
    );
  },
);
