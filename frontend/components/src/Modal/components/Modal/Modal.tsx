import classNames from 'classnames';
import noop from 'lodash/noop';
import React, {useEffect, useRef} from 'react';
import {createPortal} from 'react-dom';

import {Button} from '../../../Button';

import usePopUp from './../../../hooks/usePopUp';
import {CloseSVG} from './../../../Svg';
import useTrapFocus from './hooks/useTrapFocus';
import styles from './Modal.module.css';

export type ModalModes =
  | 'Small'
  | 'Default'
  | 'Long'
  | 'FullPage'
  | 'FullPagePanel';

type ModalProps = {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  className?: string;
  mode?: ModalModes;
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
  }, [showing]);

  const isFullPage = mode === 'FullPagePanel';

  return (
    <>
      {showing &&
        createPortal(
          <>
            <div
              className={classNames(styles.modalBackdrop, {
                [styles.isFullPage]: isFullPage,
              })}
              onClick={onHide}
              data-testid="Modal__backdrop"
              aria-hidden
            />
            <div
              className={classNames(
                animation,
                styles.modalWrapper,
                {
                  [styles.isFullPage]: isFullPage,
                },
                className,
              )}
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
