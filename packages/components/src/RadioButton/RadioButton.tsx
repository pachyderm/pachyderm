import classnames from 'classnames';
import React, {InputHTMLAttributes} from 'react';
import {useFormContext, RegisterOptions} from 'react-hook-form';

import RadioButtonLabel from './components/RadioButtonLabel';
import styles from './RadioButton.module.css';

export interface RadioButtonProps
  extends InputHTMLAttributes<HTMLInputElement> {
  name: string;
  validationOptions?: RegisterOptions;
}

const RadioButton: React.FC<RadioButtonProps> = ({
  children,
  className,
  validationOptions = {},
  ...rest
}) => {
  const {register} = useFormContext();

  return (
    <label className={classnames(styles.base, className)}>
      <div className={styles.radioButtonContainer}>
        <input
          type="radio"
          ref={register(validationOptions)}
          className={styles.input}
          {...rest}
        />

        <div className={styles.radio} />
      </div>

      {children}
    </label>
  );
};

export default Object.assign(RadioButton, {Label: RadioButtonLabel});
