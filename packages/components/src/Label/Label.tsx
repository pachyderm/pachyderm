import classNames from 'classnames';
import React, {LabelHTMLAttributes} from 'react';
import {FieldPath, FieldValues} from 'react-hook-form';

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
    <label
      htmlFor={htmlFor}
      className={classNames(styles.base, className)}
      {...rest}
    >
      {optional ? <span>{label} (optional)</span> : label}
      {children}
      {maxLength && <MaxLength htmlFor={htmlFor} maxLength={maxLength} />}
    </label>
  );
};

export default Label;
