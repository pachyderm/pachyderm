import classnames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {
  useFormContext,
  RegisterOptions,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import useRHFInputProps from 'hooks/useRHFInputProps';

import RadioButtonLabel from './components/RadioButtonLabel';
import styles from './RadioButton.module.css';

export interface RadioButtonProps
  extends InputHTMLAttributes<HTMLInputElement> {
  name: FieldPath<FieldValues>;
  validationOptions?: RegisterOptions;
}

const RadioButton: React.FC<RadioButtonProps> = ({
  children,
  className,
  validationOptions = {},
  name,
  onChange,
  onBlur,
  ...rest
}) => {
  const {register} = useFormContext();

  const {handleChange, handleBlur, ...inputProps} = useRHFInputProps({
    onChange,
    onBlur,
    registerOutput: register(name, validationOptions),
  });

  return (
    <label className={classnames(styles.base, className)}>
      <div className={styles.radioButtonContainer}>
        <input
          type="radio"
          className={styles.input}
          onChange={handleChange}
          onBlur={handleBlur}
          {...rest}
          {...inputProps}
        />

        <div className={styles.radio} />
      </div>

      {children}
    </label>
  );
};

export default Object.assign(RadioButton, {Label: RadioButtonLabel});
