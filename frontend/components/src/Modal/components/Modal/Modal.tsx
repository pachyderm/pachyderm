import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useEffect, useRef} from 'react';
import {createPortal} from 'react-dom';

import {Button} from '../../../Button';

import usePopUp from './../../../hooks/usePopUp';
import {CloseSVG} from './../../../Svg';
import useTrapFocus from './hooks/useTrapFocus';
import styles from './Modal.module.css';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  className?: string;
  small?: boolean;
  children?: React.ReactNode;
};

const Modal = ({
  children,
  show,
  onHide = noop,
  onShow = noop,
  className,
  small = false,
}: ModalProps) => {
  const {showing, animation} = usePopUp(show);
  const modalRef = useRef<HTMLDivElement>(null);

  useTrapFocus(modalRef, onHide);

  useEffect(() => {
    if (show && onShow) {
      onShow();
    }
  }, [show, onShow]);

  return (
    <>
      {showing &&
        createPortal(
          <>
            <div
              className={styles.modalBackdrop}
              onClick={onHide}
              data-testid="Modal__backdrop"
              aria-hidden
            />
            <div
              className={classNames(animation, className, styles.modalWrapper, {
                [styles.small]: small,
              })}
              ref={modalRef}
              role="dialog"
              aria-modal="true"
            >
              <div className={styles.modalDialog}>
                <div className={styles.modalContent}>
                  <Button
                    aria-label="Close"
                    data-testid="Modal__close"
                    onClick={onHide}
                    className={styles.close}
                    IconSVG={CloseSVG}
                    buttonType="ghost"
                    color="black"
                  />
                  {children}
                </div>
              </div>
            </div>
          </>,
          document.body,
        )}
    </>
  );
};

export default Modal;
