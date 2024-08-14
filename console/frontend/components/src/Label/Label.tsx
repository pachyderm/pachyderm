import classNames from 'classnames';
import React, {LabelHTMLAttributes} from 'react';
import {FieldPath, FieldValues} from 'react-hook-form';

import {FieldText} from './../Text';
import MaxLength from './components/MaxLength';
import styles from './Label.module.css';

interface LabelProps extends LabelHTMLAttributes<HTMLLabelElement> {
  optional?: boolean;
  htmlFor: FieldPath<FieldValues>;
  label: string;
  maxLength?: number;
}

const Label: React.FC<LabelProps> = ({
  children,
  className,
  htmlFor,
  label,
  maxLength,
  optional = false,
  ...rest
}) => {
  return (
    <label htmlFor={htmlFor} className={styles.base} {...rest}>
      <FieldText className={classNames(styles.content, className)}>
        {optional ? <FieldText>{label} (optional)</FieldText> : label}
        {children}
        {maxLength && <MaxLength htmlFor={htmlFor} maxLength={maxLength} />}
      </FieldText>
    </label>
  );
};

export default Label;
