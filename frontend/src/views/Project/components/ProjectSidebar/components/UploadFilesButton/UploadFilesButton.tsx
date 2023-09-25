import React from 'react';

import {Button, Link, Tooltip, UploadSVG, Icon} from '@pachyderm/components';

import useUploadFilesButton from './hooks/useUploadFilesButton';
import styles from './UploadFilesButton.module.css';

type UploadFilesButtonProps = {
  link?: boolean;
};

const UploadFilesButton = ({link}: UploadFilesButtonProps) => {
  const {fileUploadPath, disableButton, tooltipText} = useUploadFilesButton();

  if (link) {
    return !disableButton ? (
      <Link className={styles.link} to={fileUploadPath}>
        Upload files
        <Icon small color="inherit">
          <UploadSVG className={styles.linkIcon} />
        </Icon>
      </Link>
    ) : null;
  }

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
