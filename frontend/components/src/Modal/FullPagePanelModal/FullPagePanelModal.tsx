import classNames from 'classnames';
import React, {useMemo, useState} from 'react';
import BootstrapModal, {
  ModalProps as BootstrapModalProps,
} from 'react-bootstrap/Modal';

import usePopUp from '@pachyderm/components/hooks/usePopUp';

import ModalContext from './contexts/ModalContext';
import styles from './FullPagePanelModal.module.css';

export interface FullPageModalProps
  extends Omit<BootstrapModalProps, 'show' | 'onHide' | 'onShow'> {
  show: boolean;
  hideType?: 'cancel' | 'exit';
  hideLeftPanel?: boolean;
  onHide?: () => void;
  onShow?: () => void;
}

const FullPagePanelModal: React.FC<FullPageModalProps> = ({
  show,
  className,
  children,
  hideType = 'cancel',
  hideLeftPanel = false,
  onHide,
  onShow,
}) => {
  const {animation, showing} = usePopUp(show);

  const [leftOpen, setLeftOpen] = useState(false);
  const [rightOpen, setRightOpen] = useState(false);

  const modalContext = useMemo(
    () => ({
      show,
      leftOpen,
      rightOpen,
      hideLeftPanel,
      setLeftOpen,
      setRightOpen,
      hideType,
      onHide,
      onShow,
    }),
    [hideType, leftOpen, onHide, onShow, rightOpen, hideLeftPanel, show],
  );
  return (
    <ModalContext.Provider value={modalContext}>
      <BootstrapModal
        className={classNames(styles.base, animation, className)}
        show={showing}
        onHide={onHide}
        onShow={onShow}
      >
        {children}
      </BootstrapModal>
    </ModalContext.Provider>
  );
};

export default FullPagePanelModal;
