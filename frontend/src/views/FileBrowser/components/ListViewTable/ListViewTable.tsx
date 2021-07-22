import {Table} from '@pachyderm/components';
import React from 'react';

import {File} from '@graphqlTypes';

import FileTableRow from './FileTableRow';
import useListViewTable from './hooks/useListViewTable';
import styles from './ListViewTable.module.css';

type ListViewTableProps = {
  files: File[];
};

const ListViewTable: React.FC<ListViewTableProps> = ({files}) => {
  const {
    comparatorName,
    nameClick,
    sizeClick,
    dateClick,
    typeClick,
    reversed,
    tableData,
  } = useListViewTable(files);

  return (
    <Table className={styles.base} data-testid="ListViewTable__view">
      <Table.Head sticky>
        <Table.Row>
          <Table.HeaderCell
            onClick={nameClick}
            sortable={true}
            sortLabel="name"
            sortSelected={comparatorName === 'Name'}
            sortReversed={!reversed}
          >
            Name
          </Table.HeaderCell>
          <Table.HeaderCell
            onClick={sizeClick}
            sortable={true}
            sortLabel="size"
            sortSelected={comparatorName === 'Size'}
            sortReversed={!reversed}
          >
            Size
          </Table.HeaderCell>
          <Table.HeaderCell
            onClick={dateClick}
            sortable={true}
            sortLabel="date"
            sortSelected={comparatorName === 'Date'}
            sortReversed={!reversed}
          >
            Date Uploaded
          </Table.HeaderCell>
          <Table.HeaderCell
            onClick={typeClick}
            sortable={true}
            sortLabel="type"
            sortSelected={comparatorName === 'Type'}
            sortReversed={!reversed}
          >
            Type
          </Table.HeaderCell>
          <Table.HeaderCell>Actions</Table.HeaderCell>
        </Table.Row>
      </Table.Head>
      <Table.Body>
        {tableData.map((file) => (
          <FileTableRow file={file} key={file.path} />
        ))}
      </Table.Body>
    </Table>
  );
};

export default ListViewTable;
