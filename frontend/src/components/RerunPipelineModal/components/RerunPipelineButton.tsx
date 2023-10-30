import React from 'react';

import {Tooltip, Button, ActionUpdateSVG} from '@pachyderm/components';

import useRerunPipelineButton from '../hooks/useRerunPipelineButton';
import RerunPipelineModal from '../RerunPipelineModal';

type RerunPipelineButtonProps = {
  children?: React.ReactNode;
};

const RerunPipelineButton: React.FC<RerunPipelineButtonProps> = ({
  children,
}) => {
  const {pipelineId, openModal, closeModal, isOpen, disable, tooltipText} =
    useRerunPipelineButton();

  return (
    <>
      {isOpen && (
        <RerunPipelineModal
          pipelineId={pipelineId}
          show={isOpen}
          onHide={closeModal}
        />
      )}
      <Tooltip tooltipText={tooltipText()} disabled={!!children && !disable}>
        <Button
          IconSVG={ActionUpdateSVG}
          onClick={() => openModal()}
          data-testid="RerunPipelineButton__link"
          buttonType="ghost"
          aria-label="Rerun Pipeline"
          iconPosition="end"
          color="black"
          disabled={disable}
        >
          {children}
        </Button>
      </Tooltip>
    </>
  );
};

export default RerunPipelineButton;
