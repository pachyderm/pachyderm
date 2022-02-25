import {Icon, UploadSVG, Link} from '@pachyderm/components';
import React from 'react';

import useUploadFilesButton from './hooks/useUploadFilesButton';
import styles from './UploadFilesButton.module.css';

const UploadFilesButton: React.FC = () => {
  const {fileUploadPath, loading} = useUploadFilesButton();
  return (
    <Link to={fileUploadPath} small className={styles.base}>
      <Icon
        color="plum"
        className={styles.uploadSVG}
        disabled={loading}
        aria-hidden="true"
      >
        <UploadSVG />
      </Icon>
      Upload Files
    </Link>
  );
};

export default UploadFilesButton;
