import {File, FileType, FileCommitState} from '@graphqlTypes';
import classnames from 'classnames';
import React from 'react';

import {
  LegacyTable as Table,
  Group,
  AddCircleSVG,
  StatusUpdatedSVG,
  Icon,
  Tooltip,
  ButtonGroup,
  Button,
} from '@pachyderm/components';

import useFileDisplay from '../../../hooks/useFileDisplay';
import DeleteFileButton from '../../DeleteFileButton';

import styles from './FileTableRow.module.css';

type FileTableRowProps = {
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
        <ButtonGroup>
          {file.type === FileType.FILE ? (
            <>
              {previewSupported && (
                <Button buttonType="ghost" to={filePath}>
                  Preview
                </Button>
              )}
              <Button href={download} download buttonType="ghost">
                Download
              </Button>
              <DeleteFileButton file={file}>Delete</DeleteFileButton>
            </>
          ) : (
            <Button to={filePath} buttonType="ghost">
              See Files
            </Button>
          )}
          {copySupported && (
            <Button buttonType="ghost" onClick={copy} aria-label="Copy">
              Copy Path
            </Button>
          )}
        </ButtonGroup>
      </Table.DataCell>
    </Table.Row>
  );
};

export default FileTableRow;
