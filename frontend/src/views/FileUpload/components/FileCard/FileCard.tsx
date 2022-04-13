import {
  CloseSVG,
  Group,
  Icon,
  StatusCheckmarkSVG,
  StatusWarningSVG,
} from '@pachyderm/components';
import filesize from 'filesize';
import React from 'react';

import FileIcon from '@dash-frontend/components/FileIcon';

import ProgressBar from './components/ProgressBar';
import styles from './FileCard.module.css';
import useFileCard from './hooks/useFileCard';

type FileCardProps = {
  file: File;
  handleFileCancel: (index: number, success: boolean) => void;
  uploadId?: string;
  index: number;
  maxStreamIndex: number;
  onComplete: React.Dispatch<React.SetStateAction<number>>;
  onError: React.Dispatch<React.SetStateAction<string>>;
  uploadError: boolean;
};

const FileCard: React.FC<FileCardProps> = ({
  file,
  handleFileCancel,
  index,
  maxStreamIndex,
  onComplete,
  uploadId,
  onError,
  uploadError,
}) => {
  const {
    fileMajorType,
    loading,
    error,
    success,
    progress,
    cancel,
    cancelLoading,
  } = useFileCard({
    file,
    maxStreamIndex,
    index,
    onComplete,
    uploadId,
    handleFileCancel,
    onError,
    uploadError,
  });

  return (
    <Group vertical>
      <div className={styles.base}>
        <button
          type="button"
          className={styles.cancelButton}
          onClick={cancel}
          disabled={cancelLoading}
          data-testid="FileCard__cancel"
        >
          <Icon color="black" small disabled={cancelLoading}>
            <CloseSVG aria-label="remove file" />
          </Icon>
        </button>
        <Group spacing={8} align="start" className={styles.infoWrapper}>
          <FileIcon fileType={fileMajorType} />
          <div className={styles.info}>
            <h6 className={styles.name}>{file.name}</h6>
            <Group
              align="center"
              justify="between"
              className={styles.sizeGroup}
            >
              <span className={styles.size}>{filesize(file.size)}</span>
              {loading && <ProgressBar value={progress} max={100} />}
              {!error && !loading && success && (
                <Icon small data-testid="FileCard__success">
                  <StatusCheckmarkSVG aria-label="upload success" />
                </Icon>
              )}
              {error && (
                <Icon small>
                  <StatusWarningSVG aria-label="file error" />
                </Icon>
              )}
            </Group>
          </div>
        </Group>
      </div>
    </Group>
  );
};

export default FileCard;
