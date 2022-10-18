import {File, FileType, FileCommitState} from '@graphqlTypes';
import {
  CopySVG,
  Link,
  DownloadSVG,
  Group,
  Tooltip,
  Icon,
  AddCircleSVG,
  StatusUpdatedSVG,
  Button,
  ButtonGroup,
} from '@pachyderm/components';
import classnames from 'classnames';
import capitalize from 'lodash/capitalize';
import React from 'react';

import FileIcon from '../../../../components/FileIcon';
import useFileDisplay from '../../hooks/useFileDisplay';
import DeleteFileButton from '../DeleteFileButton';

import styles from './IconView.module.css';

type IconViewProps = {
  file: File;
};

const getIcon = (file: File) => {
  switch (file.commitAction) {
    case FileCommitState.UPDATED:
      return <StatusUpdatedSVG />;
    case FileCommitState.ADDED:
      return <AddCircleSVG />;
    default:
      return undefined;
  }
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
        <FileIcon fileType={fileMajorType} className={styles.fileIcon} />
        <div className={styles.fileInfo}>
          <h6 className={styles.fileName}>{fileName}</h6>
          <p className={styles.fileText}>Size: {file.sizeDisplay} </p>
        </div>
      </div>
      <Group className={styles.actions}>
        {file.commitAction && (
          <Group
            className={classnames(styles.commitStatus, {
              [styles[file.commitAction]]: file.commitAction,
            })}
            align="center"
          >
            <Icon small color="green" className={styles.commitStatusIcon}>
              {getIcon(file)}
            </Icon>
            {capitalize(file.commitAction)}
          </Group>
        )}
        <ButtonGroup>
          {copySupported && (
            <Tooltip tooltipKey={filePath} tooltipText="Copy Path">
              <Button
                buttonType="ghost"
                onClick={copy}
                aria-label="Copy"
                IconSVG={CopySVG}
              />
            </Tooltip>
          )}
          {file.type === FileType.FILE ? (
            !file.downloadDisabled && file.download ? (
              <>
                <Button
                  href={file.download}
                  download
                  aria-label={`Download ${file.path}`}
                  IconSVG={DownloadSVG}
                  buttonType="ghost"
                />
                <DeleteFileButton
                  file={file}
                  aria-label={`Delete ${file.path}`}
                />
                {previewSupported && (
                  <Button buttonType="ghost" to={filePath}>
                    Preview
                  </Button>
                )}
              </>
            ) : (
              <>
                <Tooltip
                  tooltipKey={`${filePath}download`}
                  tooltipText="This file is too large to download"
                >
                  <Button
                    IconSVG={DownloadSVG}
                    buttonType="ghost"
                    aria-label={`Download ${file.path}`}
                  />
                </Tooltip>
                <Tooltip
                  tooltipKey={`${filePath}preview`}
                  tooltipText="This file is too large to preview"
                >
                  <Button buttonType="ghost">Preview</Button>
                </Tooltip>
              </>
            )
          ) : (
            <Link to={filePath}>See Files</Link>
          )}
        </ButtonGroup>
      </Group>
    </div>
  );
};

export default IconView;
