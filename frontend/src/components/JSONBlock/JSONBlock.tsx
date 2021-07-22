import classnames from 'classnames';
import React from 'react';

import styles from './JSONBlock.module.css';

export type JSONBlockProps = React.HTMLAttributes<HTMLPreElement>;

const JSONBlock: React.FC<JSONBlockProps> = ({
  children,
  className,
  ...rest
}) => {
  return (
    <pre className={classnames(styles.base, className)} {...rest}>
      {children}
    </pre>
  );
};

export default JSONBlock;
