import React from 'react';

import {StatusWarningSVG} from './../../../Svg';
import styles from './ModalError.module.css';

const ModalError: React.FC = ({children}) => {
  return (
    <div role="status" className={styles.base}>
      <div className={styles.content}>
        <StatusWarningSVG aria-hidden className={styles.exclamation} />
        {children}
      </div>
    </div>
  );
};

export default ModalError;
