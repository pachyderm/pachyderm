import {File} from '@graphqlTypes';
import {Button, Group, Icon, TrashSVG} from '@pachyderm/components';
import React, {useCallback} from 'react';

import useFileDisplay from '../../hooks/useFileDisplay';
import DeleteFileButton from '../DeleteFileButton';

import CSVPreview from './components/CSVPreview';
import IFramePreview from './components/IFramePreview';
import JSONPreview from './components/JSONPreview';
import TextPreview from './components/TextPreview';
import WebPreview from './components/WebPreview';
import styles from './FilePreview.module.css';

type FilePreviewProps = {
  file: File;
};

const FilePreview: React.FC<FilePreviewProps> = ({file}) => {
  const {
    copy,
    copySupported,
    fileName,
    dateDisplay,
    previewSupported,
    fileType,
    fileMajorType,
  } = useFileDisplay(file);

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
          case 'yml':
            return (
              <IFramePreview downloadLink={fileLink} fileName={fileName} />
            );
          case 'txt':
          case 'jsonl':
            return <TextPreview downloadLink={fileLink} />;
          case 'html':
          case 'htm':
            return <WebPreview downloadLink={fileLink} fileName={fileName} />;
          case 'json':
            return <JSONPreview downloadLink={fileLink} />;
          case 'csv':
          case 'tsv':
          case 'tab':
            return <CSVPreview downloadLink={fileLink} />;
        }
      }
    },
    [fileMajorType, fileName, fileType, previewSupported],
  );

  return (
    <div className={styles.base}>
      <Group inline spacing={16} className={styles.fileInfo}>
        <p className={styles.fileText}>
          Uploaded: {dateDisplay} ({file.sizeDisplay})
        </p>
        {copySupported && (
          <Button
            autoWidth
            buttonType="secondary"
            onClick={copy}
            aria-label={`Copy Path ${file.path}`}
          >
            Copy Path
          </Button>
        )}
        {!file.downloadDisabled && file.download && (
          <Button
            download
            href={file.download}
            autoWidth
            aria-label={`Download ${file.path}`}
          >
            Download
          </Button>
        )}
        <DeleteFileButton file={file} aria-label={`Delete ${file.path}`}>
          <Icon color="plum">
            <TrashSVG />
          </Icon>
        </DeleteFileButton>
      </Group>
      {!file.downloadDisabled &&
        file.download &&
        getPreviewElement(file.download)}
    </div>
  );
};

export default FilePreview;
