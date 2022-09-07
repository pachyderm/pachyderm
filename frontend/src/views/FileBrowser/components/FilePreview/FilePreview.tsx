import {File} from '@graphqlTypes';
import {Button, ButtonGroup} from '@pachyderm/components';
import React, {useCallback} from 'react';

import useFileDisplay from '../../hooks/useFileDisplay';
import DeleteFileButton from '../DeleteFileButton';

import CSVPreview from './components/CSVPreview';
import IFramePreview from './components/IFramePreview';
import JSONPreview from './components/JSONPreview';
import TextPreview from './components/TextPreview';
import WebPreview from './components/WebPreview';
import YAMLPreview from './components/YAMLPreview';
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
            return (
              <IFramePreview downloadLink={fileLink} fileName={fileName} />
            );
          case 'yml':
          case 'yaml':
            return <YAMLPreview downloadLink={fileLink} />;
          case 'txt':
          case 'jsonl':
          case 'textpb':
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
      <ButtonGroup className={styles.fileInfo}>
        <p className={styles.fileText}>
          Uploaded: {dateDisplay} ({file.sizeDisplay})
        </p>
        {copySupported && (
          <Button
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
            aria-label={`Download ${file.path}`}
          >
            Download
          </Button>
        )}
        <DeleteFileButton file={file} aria-label={`Delete ${file.path}`} />
      </ButtonGroup>
      {!file.downloadDisabled &&
        file.download &&
        getPreviewElement(file.download)}
    </div>
  );
};

export default FilePreview;
