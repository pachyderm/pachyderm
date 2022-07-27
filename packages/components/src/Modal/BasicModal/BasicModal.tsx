import noop from 'lodash/noop';
import React from 'react';

import {LoadingDots} from 'LoadingDots';

import {Modal} from './../../Modal';
import styles from './BasicModal.module.css';

type BasicModalProps = {
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
  className?: string;
  hideActions?: boolean;
  errorMessage?: string;
  successMessage?: string;
  small?: boolean;
  cancelText?: string;
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
  className,
  hideActions = false,
  errorMessage = '',
  successMessage = '',
  small = false,
  cancelText,
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
      actionable={actionable}
      className={className}
      small={small}
    >
      {modalStatus ? (
        <Modal.Status status={modalStatus}>
          {errorMessage || successMessage}
        </Modal.Status>
      ) : null}

      <Modal.Header
        onHide={onHide}
        actionable={actionable}
        small={small}
        withStatus={!!modalStatus}
      >
        {headerContent}
      </Modal.Header>

      <Modal.Body>{loading ? <LoadingDots /> : children}</Modal.Body>

      {!hideActions && actionable ? (
        <Modal.Footer
          cancelTestId={cancelTestId}
          confirmTestId={confirmTestId}
          confirmText={confirmText}
          disabled={disabled}
          onConfirm={onConfirm || onHide}
          onHide={onHide}
          cancelText={cancelText}
        />
      ) : (
        <div className={styles.infoFooter} />
      )}
    </Modal>
  );
};

export default BasicModal;
