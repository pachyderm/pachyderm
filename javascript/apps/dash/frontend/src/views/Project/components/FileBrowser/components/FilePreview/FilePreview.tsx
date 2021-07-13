import {Button, Group} from '@pachyderm/components';
import React, {useCallback} from 'react';

import {File} from '@graphqlTypes';

import useFileDisplay from './../../hooks/useFileDisplay';
import CSVPreview from './components/CSVPreview';
import IFramePreview from './components/IFramePreview';
import JSONPreview from './components/JSONPreview';
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
          case 'pdf':
          case 'xml':
          case 'yml':
          case 'txt':
            return (
              <IFramePreview downloadLink={fileLink} fileName={fileName} />
            );
          case 'html':
          case 'htm':
            return <WebPreview downloadLink={fileLink} fileName={fileName} />;
          case 'json':
            return <JSONPreview downloadLink={fileLink} />;
          case 'csv':
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
          <Button autoWidth buttonType="secondary" onClick={copy}>
            Copy Path
          </Button>
        )}
        {!file.downloadDisabled && file.download && (
          <Button download href={file.download} autoWidth>
            Download
          </Button>
        )}
      </Group>
      {!file.downloadDisabled &&
        file.download &&
        getPreviewElement(file.download)}
    </div>
  );
};

export default FilePreview;
