import classNames from 'classnames';
import React from 'react';
import BootstrapModal, {
  ModalProps as BootstrapModalProps,
} from 'react-bootstrap/Modal';

import usePopUp from 'hooks/usePopUp';

import {Button} from './../../Button';
import {ExitSVG} from './../../Svg';
import styles from './FullPageModal.module.css';

export interface FullPageModalProps
  extends Omit<BootstrapModalProps, 'show' | 'onHide' | 'onShow'> {
  show: boolean;
  hideType?: 'cancel' | 'exit';
  onHide?: () => void;
  onShow?: () => void;
}

const FullPageModal: React.FC<FullPageModalProps> = ({
  show,
  className,
  children,
  hideType = 'cancel',
  onHide,
  onShow,
}) => {
  const {animation, showing} = usePopUp(show);

  return (
    <BootstrapModal
      className={classNames(styles.base, animation, className)}
      show={showing}
      onHide={onHide}
      onShow={onShow}
    >
      <BootstrapModal.Header className={styles.header}>
        {hideType === 'cancel' && (
          <Button
            autoSize
            buttonType="secondary"
            className={styles.cancelButton}
            onClick={onHide}
          >
            Cancel
          </Button>
        )}
        {hideType === 'exit' && (
          <button
            aria-label="Close"
            data-testid="FullPageModal__close"
            onClick={onHide}
            className={styles.close}
          >
            <ExitSVG className={styles.icon} />
          </button>
        )}
      </BootstrapModal.Header>
      <BootstrapModal.Body className={styles.body}>
        {children}
      </BootstrapModal.Body>
    </BootstrapModal>
  );
};

export default FullPageModal;
