import classNames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {
  useFormContext,
  RegisterOptions,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import useRHFInputProps from 'hooks/useRHFInputProps';
import {Icon} from 'Icon';

import {RadioSVG, RadioCheckedSVG} from '../Svg';

import RadioButtonLabel from './components/RadioButtonLabel';
import styles from './RadioButton.module.css';

export interface RadioButtonProps
  extends InputHTMLAttributes<HTMLInputElement> {
  name: FieldPath<FieldValues>;
  validationOptions?: RegisterOptions;
  small?: boolean;
}

const RadioButton: React.FC<RadioButtonProps> = ({
  children,
  className,
  validationOptions = {},
  name,
  onChange,
  onBlur,
  small = false,
  disabled = false,
  value,
  ...rest
}) => {
  const {register, watch} = useFormContext();

  const watchedValue = watch(name);

  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name, validationOptions),
  });

  const classes = classNames(styles.base, className, {
    [styles.disabled]: disabled,
  });

  return (
    <label className={classes} data-disabled={disabled}>
      <div className={styles.radioButtonContainer}>
        <input
          type="radio"
          className={styles.input}
          onChange={handleChange}
          onBlur={handleBlur}
          value={value}
          disabled={disabled}
          {...rest}
          {...inputProps}
        />
        <Icon small={small}>
          {watchedValue === value ? (
            <RadioCheckedSVG aria-hidden focusable={false} />
          ) : (
            <RadioSVG aria-hidden focusable={false} />
          )}
        </Icon>
      </div>

      {children}
    </label>
  );
};

export default Object.assign(RadioButton, {Label: RadioButtonLabel});
