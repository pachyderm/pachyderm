import {File} from '@graphqlTypes';
import React, {useCallback} from 'react';

import CodePreview from '@dash-frontend/components/CodePreview';
import Description from '@dash-frontend/components/Description';
import EmptyState from '@dash-frontend/components/EmptyState';
import {
  Button,
  DefaultDropdown,
  OverflowSVG,
  Group,
  ArrowLeftSVG,
  BasicModal,
  Switch,
} from '@pachyderm/components';

import useFileActions from '../../hooks/useFileActions';
import useFileDelete from '../../hooks/useFileDelete';

import CSVPreview from './components/CSVPreview';
import IFramePreview from './components/IFramePreview';
import MarkdownPreview from './components/MarkdownPreview';
import WebPreview from './components/WebPreview';
import styles from './FilePreview.module.css';

type FilePreviewProps = {
  file: File;
};

const FilePreview: React.FC<FilePreviewProps> = ({file}) => {
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
    previewSupported,
    viewSourceSupported,
    viewSource,
    toggleViewSource,
    onMenuSelect,
    iconItems,
    fileMajorType,
    fileType,
    handleBackNav,
    branchId,
  } = useFileActions(file, openDeleteModal);

  const getPreviewElement = useCallback(
    (fileLink: string) => {
      if (previewSupported) {
        switch (fileMajorType) {
          case 'image':
            return (
              <img className={styles.preview} src={fileLink} alt={fileName} />
            );
          case 'video':
            return <video controls className={styles.preview} src={fileLink} />;
          case 'audio':
            return (
              <audio className={styles.preview} controls>
                <source src={fileLink} />
              </audio>
            );
        }
        switch (fileType) {
          case 'xml':
            return (
              <IFramePreview downloadLink={fileLink} fileName={fileName} />
            );
          case 'yml':
          case 'yaml':
            return <CodePreview downloadLink={fileLink} language="yaml" />;
          case 'txt':
          case 'jsonl':
          case 'textpb':
            return <CodePreview downloadLink={fileLink} language="text" />;
          case 'html':
          case 'htm':
            return <WebPreview downloadLink={fileLink} fileName={fileName} />;
          case 'json':
            return <CodePreview downloadLink={fileLink} language="json" />;
          case 'csv':
          case 'tsv':
          case 'tab':
            return <CSVPreview downloadLink={fileLink} />;
          case 'md':
          case 'mkd':
          case 'mdwn':
          case 'mdown':
          case 'markdown':
            return viewSource ? (
              <CodePreview downloadLink={fileLink} language="markdown" />
            ) : (
              <MarkdownPreview downloadLink={fileLink} />
            );
        }
      }
    },
    [fileMajorType, fileName, fileType, previewSupported, viewSource],
  );

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
            <Description term="Size">{file.sizeDisplay}</Description>
          </div>
          <div className={styles.description}>
            <Description term="File Path">{file.path}</Description>
          </div>
        </Group>
      </div>
      <div className={styles.actionBar}>
        {viewSourceSupported ? (
          <div className={styles.actionBarItem}>
            <Switch onChange={toggleViewSource} />
            <span className={styles.actionBarViewSourceText}>View Source</span>
          </div>
        ) : null}
      </div>
      {file.download && previewSupported ? (
        getPreviewElement(file.download)
      ) : (
        <EmptyState
          error
          title="Unable to preview this file"
          message={
            !file.download
              ? 'This file is too large to preview'
              : 'This file format is not supported for file previews'
          }
        />
      )}
      {deleteModalOpen && (
        <BasicModal
          show={deleteModalOpen}
          onHide={closeModal}
          headerContent="Are you sure you want to delete this File?"
          actionable
          small
          confirmText="Delete"
          onConfirm={deleteFile}
          loading={loading}
          errorMessage={error?.message}
        >
          {file.path}
          <br />
          {`${file.repoName}@${file.commitId}`}
        </BasicModal>
      )}
    </div>
  );
};

export default FilePreview;
