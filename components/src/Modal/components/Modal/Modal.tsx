import classNames from 'classnames';
import noop from 'lodash/noop';
import React from 'react';
import BootstrapModal, {
  ModalProps as BootstrapModalProps,
} from 'react-bootstrap/Modal';

import {Button} from '../../../Button';

import usePopUp from './../../../hooks/usePopUp';
import {CloseSVG} from './../../../Svg';
import styles from './Modal.module.css';

export interface ModalProps
  extends Omit<BootstrapModalProps, 'show' | 'onHide'> {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  actionable?: boolean;
  className?: string;
  small?: boolean;
}

const Modal: React.FC<ModalProps> = ({
  children,
  show,
  onHide = noop,
  onShow = noop,
  actionable = false,
  className,
  small = false,
  ...props
}) => {
  const {showing, animation} = usePopUp(show);

  return (
    <BootstrapModal
      {...props}
      onShow={onShow}
      className={classNames(styles.base, className, animation, {
        [styles.small]: small,
      })}
      animation={false}
      show={showing}
      backdrop={actionable ? 'static' : true}
      onHide={onHide}
    >
      <Button
        aria-label="Close"
        data-testid="Modal__close"
        onClick={onHide}
        className={classNames(styles.close, {
          [styles.notActionable]: !actionable,
        })}
        IconSVG={CloseSVG}
        buttonType="ghost"
        color="black"
      />

      {children}
    </BootstrapModal>
  );
};

export default Modal;
