import classnames from 'classnames';
import React, {HTMLAttributes} from 'react';

import styles from './Badge.module.css';

const Badge: React.FC<HTMLAttributes<HTMLSpanElement>> = ({
  children,
  className,
  ...rest
}) => (
  <span className={classnames(styles.base, className)} {...rest}>
    {children}
  </span>
);

export default Badge;
