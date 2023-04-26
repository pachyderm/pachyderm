import {File} from '@graphqlTypes';
import React, {useState} from 'react';

import {Table} from '@pachyderm/components';

import FileTableRow from './FileTableRow';
import styles from './ListViewTable.module.css';

type ListViewTableProps = {
  files: File[];
};

const ListViewTable: React.FC<ListViewTableProps> = ({files}) => {
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);

  const addSelection = (filePath: string) => {
    if (selectedFiles.includes(filePath)) {
      setSelectedFiles((selectedFiles) =>
        selectedFiles.filter((file) => file !== filePath),
      );
    } else {
      setSelectedFiles((selectedFiles) => [...selectedFiles, filePath]);
    }
  };

  return (
    <div className={styles.scrolling}>
      <Table className={styles.base} data-testid="ListViewTable__view">
        <Table.Head sticky>
          <Table.Row>
            <Table.HeaderCell>File</Table.HeaderCell>
            <Table.HeaderCell>Change</Table.HeaderCell>
            <Table.HeaderCell>Size</Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Head>
        <Table.Body>
          {files.map((file) => (
            <FileTableRow
              file={file}
              key={file.path}
              selectedFiles={selectedFiles}
              addSelection={addSelection}
            />
          ))}
        </Table.Body>
      </Table>
    </div>
  );
};

export default ListViewTable;
