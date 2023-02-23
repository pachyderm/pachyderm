import React from 'react';

import {Button, UploadSVG, Tooltip} from '@pachyderm/components';

import useUploadFilesButton from './hooks/useUploadFilesButton';

const UploadFilesButton: React.FC = () => {
  const {fileUploadPath, loading} = useUploadFilesButton();
  return (
    <Tooltip tooltipKey="upload" tooltipText="Upload Files">
      <Button
        disabled={loading || !fileUploadPath}
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
