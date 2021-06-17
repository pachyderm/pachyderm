import noop from 'lodash/noop';
import React from 'react';

import {Modal} from './../../Modal';

type BasicModalProps = {
  cancelTestId?: string;
  confirmTestId?: string;
  important?: boolean;
  success?: boolean;
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  headerContent: React.ReactNode;
  onConfirm?: () => void;
  confirmText?: string;
  actionable?: boolean;
  loading?: boolean;
  disabled?: boolean;
  className?: string;
  hideActions?: boolean;
  pinTop?: boolean;
  errorMessage?: string;
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
  success = false,
  important = false,
  confirmText = 'Okay',
  actionable = false,
  loading = true,
  disabled = false,
  className,
  hideActions = false,
  pinTop = false,
  errorMessage = '',
}) => {
  return (
    <Modal
      show={show}
      onHide={onHide}
      onShow={onShow}
      actionable={actionable}
      className={className}
      pinTop={pinTop}
    >
      {errorMessage && <Modal.Error>{errorMessage}</Modal.Error>}

      <Modal.Header
        onHide={onHide}
        actionable={actionable}
        important={important}
        success={success}
        error={Boolean(errorMessage)}
        loading={loading}
      >
        {headerContent}
      </Modal.Header>

      <Modal.Body>{children}</Modal.Body>

      {!hideActions && (
        <Modal.Footer
          cancelTestId={cancelTestId}
          confirmTestId={confirmTestId}
          confirmText={confirmText}
          disabled={disabled}
          onConfirm={onConfirm || onHide}
          onHide={onHide}
        />
      )}
    </Modal>
  );
};

export default BasicModal;
