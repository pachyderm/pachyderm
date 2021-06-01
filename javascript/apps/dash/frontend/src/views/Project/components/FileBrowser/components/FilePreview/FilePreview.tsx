import {Button, Group} from '@pachyderm/components';
import React, {useCallback} from 'react';

import {File} from '@graphqlTypes';

import useFileDisplay from './../../hooks/useFileDisplay';
import styles from './FilePreview.module.css';

type FilePreviewProps = {
  file: File;
};

const FilePreview: React.FC<FilePreviewProps> = ({file}) => {
  const {copy, copySupported, fileName, dateDisplay, previewSupported} =
    useFileDisplay(file);

  const getPreviewElement = useCallback(
    (fileLink: string, fileType: string) => {
      if (previewSupported) {
        switch (fileType) {
          case 'image':
            return (
              <img
                className={styles.preview}
                src={file.download || ''}
                alt={fileName}
              />
            );
          case 'video':
            return <video controls className={styles.preview} src={fileLink} />;
        }
      }
    },
    [file.download, fileName, previewSupported],
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
        getPreviewElement(file.download, 'image')}
    </div>
  );
};

export default FilePreview;
