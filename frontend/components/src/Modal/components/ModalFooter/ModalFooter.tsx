import classnames from 'classnames';
import React, {ButtonHTMLAttributes} from 'react';
import BootstrapModalFooter from 'react-bootstrap/ModalFooter';

import {Button, ButtonGroup} from './../../../Button';
import {Icon} from './../../../Icon';
import {Link} from './../../../Link';
import {ExternalLinkSVG} from './../../../Svg';
import styles from './ModalFooter.module.css';

export interface ModalProps {
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
  infoLink?: string;
  infoLinkText?: string;
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
  infoLink,
  infoLinkText = 'Info',
}) => {
  return (
    <BootstrapModalFooter className={classnames(styles.base, className)}>
      {infoLink && (
        <Link externalLink inline to={infoLink} className={styles.footerLink}>
          {infoLinkText}
          <Icon small color="plum">
            <ExternalLinkSVG />
          </Icon>
        </Link>
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
