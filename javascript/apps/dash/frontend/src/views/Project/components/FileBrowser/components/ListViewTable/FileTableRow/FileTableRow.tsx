import {ButtonLink, Link, Table, Group} from '@pachyderm/components';
import React from 'react';

import {File, FileType} from '@graphqlTypes';

import useFileDisplay from './../../../hooks/useFileDisplay';
import styles from './FileTableRow.module.css';

type FileTableRowProps = {
  file: File;
};

const FileTableRow: React.FC<FileTableRowProps> = ({file}) => {
  const {
    copy,
    copySupported,
    fileName,
    dateDisplay,
    filePath,
    fileType,
    previewSupported,
  } = useFileDisplay(file);

  const download =
    !file.downloadDisabled && file.download ? file.download : undefined;

  return (
    <Table.Row className={styles.base}>
      <Table.DataCell>{fileName}</Table.DataCell>
      <Table.DataCell>{file.sizeDisplay}</Table.DataCell>
      <Table.DataCell>{dateDisplay}</Table.DataCell>
      <Table.DataCell>{fileType}</Table.DataCell>
      <Table.DataCell>
        <Group spacing={16}>
          {file.type === FileType.FILE ? (
            <>
              {previewSupported && <Link to={filePath}>Preview</Link>}
              <Link to={download} download>
                Download
              </Link>
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
