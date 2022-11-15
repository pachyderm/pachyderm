import classNames from 'classnames';
import React from 'react';

import styles from './HelperText.module.css';

const HelperText: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  children,
  className,
}) => {
  return <span className={classNames(styles.base, className)}>{children}</span>;
};

export default HelperText;
