import classnames from 'classnames';
import React, {ButtonHTMLAttributes} from 'react';
import BootstrapModalFooter from 'react-bootstrap/ModalFooter';

import {Button, ButtonGroup} from './../../../Button';
import styles from './ModalFooter.module.css';

export interface ModalProps {
  children?: React.ReactNode;
  cancelTestId?: string;
  confirmText: string;
  confirmTestId?: string;
  onConfirm?: () => void;
  onHide: () => void;
  buttonType?: ButtonHTMLAttributes<HTMLButtonElement>['type'];
  disabled?: boolean;
  className?: string;
  cancelText?: string;
  hideConfirm?: boolean;
  footerContent?: JSX.Element;
}

const ModalFooter: React.FC<ModalProps> = ({
  cancelTestId = '',
  confirmTestId = '',
  confirmText,
  onHide,
  onConfirm,
  buttonType = 'button',
  disabled = false,
  className = '',
  cancelText = 'Cancel',
  hideConfirm = false,
  footerContent,
}) => {
  return (
    <BootstrapModalFooter className={classnames(styles.base, className)}>
      {footerContent && (
        <div className={styles.footerContent}>{footerContent}</div>
      )}
      <ButtonGroup>
        <Button
          data-testid={cancelTestId || 'ModalFooter__cancel'}
          onClick={onHide}
          type="button"
          buttonType={hideConfirm ? 'primary' : 'ghost'}
        >
          {cancelText}
        </Button>
        {!hideConfirm && (
          <Button
            data-testid={confirmTestId || 'ModalFooter__confirm'}
            onClick={onConfirm}
            type={buttonType}
            disabled={disabled}
          >
            {confirmText}
          </Button>
        )}
      </ButtonGroup>
    </BootstrapModalFooter>
  );
};

export default ModalFooter;
