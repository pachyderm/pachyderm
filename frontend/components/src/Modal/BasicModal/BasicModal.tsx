import noop from 'lodash/noop';
import React from 'react';

import {LoadingDots} from '@pachyderm/components';

import {ModalModes} from '../components/Modal/Modal';

import {Modal} from './../../Modal';
import styles from './BasicModal.module.css';

type BasicModalProps = {
  children?: React.ReactNode;
  cancelTestId?: string;
  confirmTestId?: string;
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  headerContent: React.ReactNode;
  onConfirm?: () => void;
  confirmText?: string;
  actionable?: boolean;
  loading?: boolean;
  updating?: boolean;
  disabled?: boolean;
  flexBody?: boolean;
  className?: string;
  hideActions?: boolean;
  hideConfirm?: boolean;
  errorMessage?: string;
  successMessage?: string;
  mode?: ModalModes;
  cancelText?: string;
  footerContent?: JSX.Element;
};

const BasicModal: React.FC<BasicModalProps> = ({
  cancelTestId,
  confirmTestId,
  children,
  show,
  onHide = noop,
  onShow = noop,
  headerContent,
  onConfirm,
  confirmText = 'Okay',
  actionable = false,
  loading = true,
  updating = false,
  disabled = false,
  flexBody = false,
  className,
  hideActions = false,
  errorMessage = '',
  successMessage = '',
  hideConfirm = false,
  mode = 'Default',
  cancelText,
  footerContent,
}) => {
  const modalStatus =
    (updating && 'updating') ||
    (errorMessage && 'error') ||
    (successMessage && 'success');
  return (
    <Modal
      show={show}
      onHide={onHide}
      onShow={onShow}
      className={className}
      mode={mode}
    >
      {modalStatus ? (
        <Modal.Status status={modalStatus}>
          {errorMessage || successMessage}
        </Modal.Status>
      ) : null}

      <Modal.Header small={mode === 'Small'} withStatus={!!modalStatus}>
        {headerContent}
      </Modal.Header>

      <Modal.Body className={flexBody && styles.flex}>
        {loading ? <LoadingDots /> : children}
      </Modal.Body>

      {!hideActions && actionable ? (
        <Modal.Footer
          cancelTestId={cancelTestId}
          confirmTestId={confirmTestId}
          confirmText={confirmText}
          disabled={disabled}
          onConfirm={onConfirm || onHide}
          onHide={onHide}
          cancelText={cancelText}
          hideConfirm={hideConfirm}
          footerContent={footerContent}
        />
      ) : (
        <div className={styles.infoFooter} />
      )}
    </Modal>
  );
};

export default BasicModal;
