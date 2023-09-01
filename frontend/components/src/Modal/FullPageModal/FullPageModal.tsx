import classNames from 'classnames';
import React from 'react';

import {Modal} from '@pachyderm/components';

import styles from './FullPageModal.module.css';

type FullPageModalProps = {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  className?: string;
  children?: React.ReactNode;
};

const FullPageModal: React.FC<FullPageModalProps> = ({
  show,
  className,
  children,
  onHide,
  onShow,
}) => {
  return (
    <Modal
      className={classNames(styles.base, className)}
      show={show}
      onHide={onHide}
      onShow={onShow}
      mode="FullPage"
    >
      <Modal.Body className={styles.body}>{children}</Modal.Body>
    </Modal>
  );
};

export default FullPageModal;
