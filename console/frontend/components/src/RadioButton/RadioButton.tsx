import classNames from 'classnames';
import React, {InputHTMLAttributes, forwardRef} from 'react';
import {
  useFormContext,
  RegisterOptions,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import {Icon} from '@pachyderm/components';
import useRHFInputProps from '@pachyderm/components/hooks/useRHFInputProps';

import {RadioSVG, RadioCheckedSVG} from '../Svg';

import RadioButtonLabel from './components/RadioButtonLabel';
import styles from './RadioButton.module.css';

export interface RadioButtonProps
  extends InputHTMLAttributes<HTMLInputElement> {
  name: FieldPath<FieldValues>;
  validationOptions?: RegisterOptions;
  small?: boolean;
}

export interface PureRadioButtonProps
  extends InputHTMLAttributes<HTMLInputElement> {
  selected: boolean;
  small?: boolean;
}

const RadioButton: React.FC<RadioButtonProps> = ({
  children,
  validationOptions = {},
  name,
  onChange,
  onBlur,
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

  return (
    <PureRadioButton
      onChange={handleChange}
      onBlur={handleBlur}
      value={value}
      selected={watchedValue === value}
      {...rest}
      {...inputProps}
    >
      {children}
    </PureRadioButton>
  );
};

const PureRadioButton = forwardRef<HTMLInputElement, PureRadioButtonProps>(
  function PureRadioButtonForwardRef(
    {
      className,
      selected,
      onChange,
      onBlur,
      small = false,
      disabled = false,
      children,
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
            type="radio"
            className={styles.input}
            disabled={disabled}
            onChange={onChange}
            onBlur={onBlur}
            ref={ref}
            {...rest}
          />
          <Icon small={small}>
            {selected ? (
              <RadioCheckedSVG aria-hidden focusable={false} />
            ) : (
              <RadioSVG aria-hidden focusable={false} />
            )}
          </Icon>
        </div>
        {children}
      </label>
    );
  },
);

export default Object.assign(RadioButton, {
  Label: RadioButtonLabel,
  Pure: PureRadioButton,
});
