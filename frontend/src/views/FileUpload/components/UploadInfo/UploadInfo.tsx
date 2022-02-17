import {
  Group,
  Icon,
  SpinnerSVG,
  CloseSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
  ElephantEmptyState,
} from '@pachyderm/components';
import filesize from 'filesize';
import React, {useMemo} from 'react';
import {FieldError} from 'react-hook-form';

import FileIcon from '@dash-frontend/components/FileIcon';
import getFileMajorType from '@dash-frontend/lib/getFileMajorType';

import styles from './UploadInfo.module.css';

type UploadInfoProps = {
  file: File | null;
  loading: boolean;
  uploadError?: string;
  fileError?: FieldError;
  success: boolean;
  handleFileCancel: () => void;
};

const UploadInfo: React.FC<UploadInfoProps> = ({
  file,
  loading,
  uploadError,
  fileError,
  success,
  handleFileCancel,
}) => {
  const fileType = useMemo(() => {
    return (file?.name || '')
      .slice((file?.name || '').lastIndexOf('.') + 1)
      .toLowerCase();
  }, [file?.name]);

  const fileMajorType = getFileMajorType(fileType);

  return (
    <div className={styles.base}>
      {file ? (
        <Group vertical>
          <div className={styles.file}>
            <Group spacing={8} align="start" className={styles.info}>
              <FileIcon fileType={fileMajorType} />
              <Group spacing={8} vertical>
                <span className={styles.name}>{file.name}</span>
                <span className={styles.size}>{filesize(file.size)}</span>
              </Group>
            </Group>
            <Group vertical align="center" justify="between">
              <button
                type="button"
                className={styles.cancelButton}
                onClick={handleFileCancel}
                disabled={loading}
                data-testid="UploadInfo__cancel"
              >
                <Icon color="black" small disabled={loading}>
                  <CloseSVG aria-label="remove file" />
                </Icon>
              </button>
              {loading && (
                <Icon color="grey" small data-testid="UploadInfo__loading">
                  <SpinnerSVG aria-label="loading" />
                </Icon>
              )}
              {!(uploadError || fileError) && !loading && success && (
                <Icon
                  small
                  className={styles.success}
                  data-testid="UploadInfo__success"
                >
                  <StatusCheckmarkSVG aria-label="upload success" />
                </Icon>
              )}
              {(uploadError || fileError) && (
                <Icon small>
                  <StatusWarningSVG aria-label="file error" />
                </Icon>
              )}
            </Group>
          </div>
          <div className={styles.status}>
            {fileError && (
              <span className={styles.error} id="file-error">
                {fileError.message}
              </span>
            )}
            {loading && (
              <span className={styles.loading}>
                Please keep your view open during uploading…
              </span>
            )}
          </div>
        </Group>
      ) : (
        <Group
          vertical
          spacing={32}
          align="center"
          className={styles.emptyState}
        >
          <ElephantEmptyState />
          <Group
            vertical
            spacing={16}
            align="center"
            className={styles.emptyStateText}
          >
            <span>Let&apos;s Start:)</span>
            <span>Upload your file on the left!</span>
          </Group>
        </Group>
      )}
    </div>
  );
};

export default UploadInfo;
