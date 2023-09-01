import React from 'react';

import {Button, UploadSVG, Tooltip} from '@pachyderm/components';

import useUploadFilesButton from './hooks/useUploadFilesButton';

const UploadFilesButton: React.FC = () => {
  const {fileUploadPath, disableButton, tooltipText} = useUploadFilesButton();
  return (
    <Tooltip tooltipText={tooltipText}>
      <Button
        disabled={disableButton}
        to={fileUploadPath}
        data-testid="UploadFilesButton__button"
        buttonType="ghost"
        IconSVG={UploadSVG}
        aria-label="Upload Files"
      />
    </Tooltip>
  );
};

export default UploadFilesButton;
