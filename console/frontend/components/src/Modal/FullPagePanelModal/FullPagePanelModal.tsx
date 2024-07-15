import React, {useMemo, useState} from 'react';

import {Modal} from '@pachyderm/components';

import ModalContext from './contexts/ModalContext';

type FullPageModalProps = {
  show: boolean;
  hideLeftPanel?: boolean;
  onHide?: () => void;
  onShow?: () => void;
  className?: string;
  children?: React.ReactNode;
};

const FullPagePanelModal: React.FC<FullPageModalProps> = ({
  show,
  className,
  children,
  hideLeftPanel = false,
  onHide,
  onShow,
}) => {
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
      onHide,
      onShow,
    }),
    [leftOpen, onHide, onShow, rightOpen, hideLeftPanel, show],
  );
  return (
    <ModalContext.Provider value={modalContext}>
      <Modal
        className={className}
        show={show}
        onHide={onHide}
        onShow={onShow}
        mode="FullPagePanel"
        noCloseButton
      >
        {children}
      </Modal>
    </ModalContext.Provider>
  );
};

export default FullPagePanelModal;
