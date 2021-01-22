import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './RadioButtonLabel.module.css';

export interface RadioButtonLabelProps extends HTMLAttributes<HTMLSpanElement> {
  small?: boolean;
}

const RadioButtonLabel: React.FC<RadioButtonLabelProps> = ({
  children,
  className,
  small,
  ...rest
}) => (
  <span
    className={classnames(styles.base, className, {[styles.small]: small})}
    {...rest}
  >
    {children}
  </span>
);

export default RadioButtonLabel;
