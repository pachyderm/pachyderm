import React from 'react';

import Description from '@dash-frontend/components/Description';
import {FileInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {getDownloadLink} from '@dash-frontend/lib/fileUtils';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {
  Button,
  DefaultDropdown,
  OverflowSVG,
  Group,
  ArrowLeftSVG,
  BasicModal,
} from '@pachyderm/components';

import useFileActions from '../../hooks/useFileActions';
import useFileDelete from '../../hooks/useFileDelete';

import FilePreviewContent from './components/FilePreviewContent';
import styles from './FilePreview.module.css';

type FilePreviewProps = {
  file: FileInfo;
};

const FilePreview = ({file}: FilePreviewProps) => {
  const {
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteFile,
    loading,
    error,
  } = useFileDelete(file);
  const {
    fileName,
    viewSource,
    toggleViewSource,
    onMenuSelect,
    iconItems,
    fileType,
    handleBackNav,
    branchId,
  } = useFileActions(file, openDeleteModal);

  return (
    <div className={styles.base}>
      <div className={styles.header}>
        <div className={styles.title}>
          <h6>{fileName}</h6>
          <Group spacing={16}>
            <Button
              buttonType="secondary"
              IconSVG={ArrowLeftSVG}
              onClick={handleBackNav}
            >
              Back to file list
            </Button>

            <DefaultDropdown
              items={iconItems}
              onSelect={onMenuSelect}
              storeSelected
              buttonOpts={{
                hideChevron: true,
                IconSVG: OverflowSVG,
                buttonType: 'ghost',
              }}
              menuOpts={{pin: 'right', className: styles.dropdown}}
            />
          </Group>
        </div>
        <Group
          spacing={64}
          className={styles.metaData}
          aria-label="file metadata"
        >
          <div className={styles.description}>
            <Description term="Branch">{branchId}</Description>
          </div>
          <div className={styles.description}>
            <Description term="File Type">{fileType}</Description>
          </div>
          <div className={styles.description}>
            <Description term="Size">
              {formatBytes(file.sizeBytes || 0)}
            </Description>
          </div>
          <div className={styles.description}>
            <Description term="File Path" className={styles.filePath}>
              {file.file?.path}
            </Description>
          </div>
        </Group>
      </div>

      <FilePreviewContent
        download={getDownloadLink(file)}
        path={file.file?.path || ''}
        viewSource={viewSource}
        toggleViewSource={toggleViewSource}
      />

      {deleteModalOpen && (
        <BasicModal
          show={deleteModalOpen}
          onHide={closeModal}
          headerContent="Are you sure you want to delete this File?"
          actionable
          mode="Small"
          confirmText="Delete"
          onConfirm={deleteFile}
          loading={loading}
          errorMessage={error}
        >
          {file.file?.path}
          <br />
          {`${file.file?.commit?.repo?.name}@${file.file?.commit?.id}`}
        </BasicModal>
      )}
    </div>
  );
};

export default FilePreview;
