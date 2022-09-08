import React from 'react';

import styles from './ModalBody.module.css';

const ModalBody: React.FC = ({children}) => {
  return <div className={styles.base}>{children}</div>;
};

export default ModalBody;
