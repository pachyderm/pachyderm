import classnames from 'classnames';
import React from 'react';

import usePanelModal from '../../hooks/usePanelModal';

import styles from './ModalBody.module.css';

const ModalBody = ({children}: {children?: React.ReactNode}) => {
  const {leftOpen, hideLeftPanel} = usePanelModal();
  return (
    <div
      className={classnames(styles.base, {
        [styles.leftPanel]: leftOpen,
        [styles.hideLeftPanel]: hideLeftPanel,
      })}
    >
      {children}
    </div>
  );
};

export default ModalBody;
