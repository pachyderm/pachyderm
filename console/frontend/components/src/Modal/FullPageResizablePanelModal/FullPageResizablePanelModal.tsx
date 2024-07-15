import React from 'react';
import {PanelGroup} from 'react-resizable-panels';

import {Modal} from '@pachyderm/components';

type FullPageModalProps = {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
  className?: string;
  children?: React.ReactNode;
  autoSaveId?: string;
};

const FullPagePanelModal: React.FC<FullPageModalProps> = ({
  show,
  className,
  children,
  onHide,
  onShow,
  autoSaveId,
}) => {
  return (
    <Modal
      className={className}
      show={show}
      onHide={onHide}
      onShow={onShow}
      mode="FullPagePanel"
      noCloseButton
    >
      <PanelGroup direction="horizontal" autoSaveId={autoSaveId}>
        {children}
      </PanelGroup>
    </Modal>
  );
};

export default FullPagePanelModal;
