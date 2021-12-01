import {File, FileType} from '@graphqlTypes';
import {
  CopySVG,
  ButtonLink,
  Link,
  DownloadSVG,
  Group,
  Tooltip,
} from '@pachyderm/components';
import React from 'react';

import useFileDisplay from '../../hooks/useFileDisplay';

import FileIcon from './components/FileIcon';
import styles from './IconView.module.css';

type IconViewProps = {
  file: File;
};

const IconView: React.FC<IconViewProps> = ({file}) => {
  const {
    copy,
    copySupported,
    fileName,
    filePath,
    fileMajorType,
    previewSupported,
  } = useFileDisplay(file);

  return (
    <div className={styles.base}>
      <div className={styles.content}>
        <FileIcon fileType={fileMajorType} />
        <div className={styles.fileInfo}>
          <h4 className={styles.fileName}>{fileName}</h4>
          <p className={styles.fileText}>Size: {file.sizeDisplay} </p>
        </div>
      </div>
      <Group spacing={16} className={styles.actions} align="center">
        {copySupported && (
          <Tooltip tooltipKey={filePath} tooltipText="Copy Path">
            <ButtonLink onClick={copy} aria-label="Copy">
              <CopySVG className={styles.actionIcon} />
            </ButtonLink>
          </Tooltip>
        )}
        {file.type === FileType.FILE ? (
          !file.downloadDisabled && file.download ? (
            <>
              <Link to={file.download} download>
                <DownloadSVG className={styles.actionIcon} />
              </Link>
              {previewSupported && <Link to={filePath}>Preview</Link>}
            </>
          ) : (
            <>
              <Tooltip
                tooltipKey={`${filePath}download`}
                tooltipText="This file is too large to download"
              >
                <Link>
                  <DownloadSVG className={styles.actionIcon} />
                </Link>
              </Tooltip>
              <Tooltip
                tooltipKey={`${filePath}preview`}
                tooltipText="This file is too large to preview"
              >
                <Link>Preview</Link>
              </Tooltip>
            </>
          )
        ) : (
          <Link to={filePath}>See Files</Link>
        )}
      </Group>
    </div>
  );
};

export default IconView;
