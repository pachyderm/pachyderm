import classnames from 'classnames';
import React, {ButtonHTMLAttributes} from 'react';
import BootstrapModalFooter from 'react-bootstrap/ModalFooter';

import {Button} from './../../../Button';
import {ButtonLink} from './../../../ButtonLink';
import {Group} from './../../../Group';
import styles from './ModalFooter.module.css';

export interface ModalProps {
  cancelTestId?: string;
  confirmText: string;
  confirmTestId?: string;
  onConfirm: () => void;
  onHide: () => void;
  buttonType?: ButtonHTMLAttributes<HTMLButtonElement>['type'];
  disabled?: boolean;
  className?: string;
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
}) => {
  return (
    <BootstrapModalFooter className={classnames(styles.base, className)}>
      <Group spacing={32}>
        <ButtonLink
          data-testid={cancelTestId || 'ModalFooter__cancel'}
          onClick={onHide}
          type="button"
        >
          Cancel
        </ButtonLink>
        <Button
          data-testid={confirmTestId || 'ModalFooter__confirm'}
          onClick={onConfirm}
          type={buttonType}
          disabled={disabled}
        >
          {confirmText}
        </Button>
      </Group>
    </BootstrapModalFooter>
  );
};

export default ModalFooter;
