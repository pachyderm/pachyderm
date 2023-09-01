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
  mode?: 'Small' | 'Default' | 'FullPage' | 'FullPagePanel';
  children?: React.ReactNode;
  noCloseButton?: boolean;
};

const Modal = ({
  children,
  show,
  onHide = noop,
  onShow = noop,
  className,
  mode = 'Default',
  noCloseButton = false,
}: ModalProps) => {
  const {showing, animation} = usePopUp(show);
  const modalRef = useRef<HTMLDivElement>(null);

  useTrapFocus(modalRef, onHide);

  useEffect(() => {
    if (show && onShow) {
      onShow();
    }
  }, [show, onShow]);

  useEffect(() => {
    if (showing) {
      document.getElementById('root')?.setAttribute('aria-hidden', 'true');
    }

    return () =>
      document.getElementById('root')?.removeAttribute('aria-hidden');
  });

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
              className={classNames(animation, styles.modalWrapper, className)}
              ref={modalRef}
              role="dialog"
              aria-modal="true"
            >
              <div className={styles[`modalDialog${mode}`]}>
                <div className={styles[`modalContent${mode}`]}>
                  {!noCloseButton && (
                    <Button
                      aria-label="Close"
                      data-testid="Modal__close"
                      onClick={onHide}
                      className={styles.close}
                      IconSVG={CloseSVG}
                      buttonType="ghost"
                      color="black"
                    />
                  )}
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
