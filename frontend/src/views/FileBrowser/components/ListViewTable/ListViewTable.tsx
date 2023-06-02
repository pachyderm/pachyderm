import {File} from '@graphqlTypes';
import React from 'react';

import {
  Table,
  Button,
  TrashSVG,
  BasicModal,
  DownloadSVG,
  Group,
  Tooltip,
} from '@pachyderm/components';

import FileTableRow from './FileTableRow';
import useListViewTable from './hooks/useListViewTable';
import styles from './ListViewTable.module.css';

type ListViewTableProps = {
  files: File[];
};

const ListViewTable: React.FC<ListViewTableProps> = ({files}) => {
  const {
    repoId,
    branchId,
    selectedFiles,
    addSelection,
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteDisabled,
    deleteFiles,
    deleteLoading,
    deleteError,
    downloadSelected,
    downloadDisabled,
    proxyEnabled,
  } = useListViewTable();

  return (
    <>
      <Group spacing={8} className={styles.headerButtons}>
        <Button
          IconSVG={TrashSVG}
          onClick={openDeleteModal}
          disabled={deleteDisabled}
          buttonType="ghost"
          color="black"
          aria-label="Delete selected items"
        />

        <Tooltip
          tooltipKey="delete"
          tooltipText={
            'Enable proxy to download multiple files at once. This feature is not available with your current configuration.'
          }
          disabled={proxyEnabled}
        >
          <span>
            <Button
              IconSVG={DownloadSVG}
              onClick={downloadSelected}
              disabled={!proxyEnabled || downloadDisabled}
              buttonType="ghost"
              color="black"
              aria-label="Download selected items"
              className={!proxyEnabled && styles.disabledButton}
            />
          </span>
        </Tooltip>
      </Group>
      <div className={styles.scrolling}>
        <Table className={styles.tableBase} data-testid="ListViewTable__view">
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
      {deleteModalOpen && (
        <BasicModal
          show={deleteModalOpen}
          onHide={closeModal}
          headerContent={`Are you sure you want to delete the selected items from ${repoId}@${branchId}?`}
          actionable
          small
          confirmText="Delete"
          onConfirm={deleteFiles}
          loading={deleteLoading}
          errorMessage={deleteError?.message}
        >
          <ul className={styles.modalList}>
            {selectedFiles.map((file) => (
              <li key={file}>{file}</li>
            ))}
          </ul>
        </BasicModal>
      )}
    </>
  );
};

export default ListViewTable;
