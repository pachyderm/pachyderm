import React from 'react';

import {EditSVG, Tooltip, Button} from '@pachyderm/components';

import useUpdatePipelineButton from './hooks/useUpdatePipelineButton';

const UpdatePipelineButton: React.FC = () => {
  const {tooltipText, disableButton, handleClick} = useUpdatePipelineButton();

  return (
    <Tooltip tooltipText={tooltipText}>
      <span>
        <Button
          IconSVG={EditSVG}
          disabled={disableButton}
          onClick={handleClick}
          buttonType="ghost"
          aria-label="Update Pipeline"
        />
      </span>
    </Tooltip>
  );
};

export default UpdatePipelineButton;
