import React from 'react';

import {Button, UploadSVG, Tooltip} from '@pachyderm/components';

import useUploadFilesButton from './hooks/useUploadFilesButton';
import styles from './UploadFilesButton.module.css';

const UploadFilesButton: React.FC = () => {
  const {fileUploadPath, disableButton, tooltipText} = useUploadFilesButton();
  return (
    <Tooltip tooltipKey="upload" tooltipText={tooltipText} size="large">
      <span>
        <Button
          disabled={disableButton}
          to={fileUploadPath}
          data-testid="UploadFilesButton__button"
          buttonType="ghost"
          IconSVG={UploadSVG}
          aria-label="Upload Files"
          className={disableButton && styles.pointerEventsNone}
        />
      </span>
    </Tooltip>
  );
};

export default UploadFilesButton;
