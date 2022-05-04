import {Button, UploadSVG} from '@pachyderm/components';
import React from 'react';

import useUploadFilesButton from './hooks/useUploadFilesButton';

const UploadFilesButton: React.FC = () => {
  const {fileUploadPath, loading} = useUploadFilesButton();
  return (
    <Button
      buttonType="ghost"
      to={fileUploadPath}
      IconSVG={UploadSVG}
      disabled={loading}
    >
      Upload Files
    </Button>
  );
};

export default UploadFilesButton;
