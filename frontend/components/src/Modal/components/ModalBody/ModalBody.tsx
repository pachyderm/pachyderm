import React from 'react';
import BootstrapModalBody from 'react-bootstrap/ModalBody';

import styles from './ModalBody.module.css';

const ModalBody = ({children}: {children?: React.ReactNode}) => {
  return (
    <BootstrapModalBody className={styles.base}>{children}</BootstrapModalBody>
  );
};

export default ModalBody;
