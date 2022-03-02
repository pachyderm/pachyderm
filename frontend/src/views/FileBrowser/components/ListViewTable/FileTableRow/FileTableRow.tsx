import {File, FileType, FileCommitState} from '@graphqlTypes';
import {
  ButtonLink,
  Link,
  Table,
  Group,
  AddCircleSVG,
  UpdatedCircleSVG,
  Icon,
  Tooltip,
} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';

import useFileDisplay from '../../../hooks/useFileDisplay';
import DeleteFileButton from '../../DeleteFileButton';

import styles from './FileTableRow.module.css';

type FileTableRowProps = {
  file: File;
};

const getIcon = (file: File) => {
  switch (file.commitAction) {
    case FileCommitState.UPDATED:
      return <UpdatedCircleSVG />;
    case FileCommitState.ADDED:
      return <AddCircleSVG />;
    default:
      return undefined;
  }
};

const FileTableRow: React.FC<FileTableRowProps> = ({file}) => {
  const {copy, copySupported, fileName, filePath, fileType, previewSupported} =
    useFileDisplay(file);

  const download =
    !file.downloadDisabled && file.download ? file.download : undefined;

  return (
    <Table.Row>
      <Table.DataCell>
        <Group spacing={8}>
          {file.commitAction && (
            <Tooltip
              tooltipKey="File Diffs"
              tooltipText={`This file was ${file.commitAction?.toLowerCase()} in this commit`}
            >
              <Icon
                small
                color="green"
                className={classnames(styles.commitIcon, {
                  [styles[file.commitAction]]: file.commitAction,
                })}
              >
                {getIcon(file)}
              </Icon>
            </Tooltip>
          )}
          {fileName}
        </Group>
      </Table.DataCell>
      <Table.DataCell>{file.sizeDisplay}</Table.DataCell>
      <Table.DataCell>{fileType}</Table.DataCell>
      <Table.DataCell>
        <Group spacing={16}>
          {file.type === FileType.FILE ? (
            <>
              {previewSupported && <Link to={filePath}>Preview</Link>}
              <Link to={download} download>
                Download
              </Link>
              <DeleteFileButton file={file}>Delete</DeleteFileButton>
            </>
          ) : (
            <Link to={filePath}>See Files</Link>
          )}
          {copySupported && (
            <ButtonLink onClick={copy} aria-label="Copy">
              Copy Path
            </ButtonLink>
          )}
        </Group>
      </Table.DataCell>
    </Table.Row>
  );
};

export default FileTableRow;
