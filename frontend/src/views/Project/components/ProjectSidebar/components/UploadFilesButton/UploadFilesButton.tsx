import React from 'react';

import {Button, UploadSVG} from '@pachyderm/components';

import useUploadFilesButton from './hooks/useUploadFilesButton';

const UploadFilesButton: React.FC = () => {
  const {fileUploadPath, loading} = useUploadFilesButton();
  return (
    <Button
      buttonType="ghost"
      to={fileUploadPath}
      IconSVG={UploadSVG}
      aria-label="Upload Files"
      disabled={loading || !fileUploadPath}
    >
      Upload Files
    </Button>
  );
};

export default UploadFilesButton;
