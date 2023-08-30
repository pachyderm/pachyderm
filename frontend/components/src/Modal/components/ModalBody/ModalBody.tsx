import classNames from 'classnames';
import React from 'react';

import styles from './ModalBody.module.css';

const ModalBody = ({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) => {
  return <div className={classNames(styles.base, className)}>{children}</div>;
};

export default ModalBody;
