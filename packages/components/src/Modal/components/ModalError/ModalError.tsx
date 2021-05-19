import React from 'react';

import {ExclamationErrorSVG} from './../../../Svg';
import styles from './ModalError.module.css';

const ModalError: React.FC = ({children}) => {
  return (
    <div role="status" className={styles.base}>
      <div className={styles.content}>
        <ExclamationErrorSVG aria-hidden className={styles.exclamation} />

        {children}
      </div>
    </div>
  );
};

export default ModalError;
