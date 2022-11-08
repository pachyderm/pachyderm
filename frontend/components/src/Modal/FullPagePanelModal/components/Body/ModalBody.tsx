import classnames from 'classnames';
import React from 'react';

import usePanelModal from '../../hooks/usePanelModal';

import styles from './ModalBody.module.css';

const ModalBody: React.FC = ({children}) => {
  const {leftOpen} = usePanelModal();
  return (
    <div className={classnames(styles.base, {[styles.leftPanel]: leftOpen})}>
      {children}
    </div>
  );
};

export default ModalBody;
