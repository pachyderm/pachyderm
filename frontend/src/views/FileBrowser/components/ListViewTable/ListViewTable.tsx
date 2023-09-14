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

const SELECT_FILES_DELETE = 'Select one or more files to multi-delete files';
const SELECT_FILES_DOWNLOAD =
  'Select one or more files to multi-download files';
const OUTPUT_REPO = 'You cannot delete files in an output repo';
const PROXY_NEEDED =
  'Enable proxy to download multiple files at once. This feature is not available with your current configuration.';

const ListViewTable: React.FC<ListViewTableProps> = ({files}) => {
  const {
    repoId,
    repoLoading,
    branchId,
    selectedFiles,
    addSelection,
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteFiles,
    deleteLoading,
    deleteError,
    downloadSelected,
    proxyEnabled,
    noFilesSelected,
    isOutputRepo,
  } = useListViewTable();

  const downloadDisabled = !proxyEnabled || noFilesSelected;
  const deleteDisabled = isOutputRepo || repoLoading || noFilesSelected;

  return (
    <>
      <Group spacing={8} className={styles.headerButtons}>
        <Tooltip
          tooltipText={isOutputRepo ? OUTPUT_REPO : SELECT_FILES_DELETE}
          disabled={!deleteDisabled}
        >
          <span>
            <Button
              IconSVG={TrashSVG}
              onClick={openDeleteModal}
              disabled={deleteDisabled}
              buttonType="ghost"
              color="black"
              aria-label="Delete selected items"
              className={deleteDisabled && styles.disabledButton}
            />
          </span>
        </Tooltip>

        <Tooltip
          tooltipText={!proxyEnabled ? PROXY_NEEDED : SELECT_FILES_DOWNLOAD}
          disabled={!downloadDisabled}
        >
          <span>
            <Button
              IconSVG={DownloadSVG}
              onClick={downloadSelected}
              disabled={downloadDisabled}
              buttonType="ghost"
              color="black"
              aria-label="Download selected items"
              className={downloadDisabled && styles.disabledButton}
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
          mode="Small"
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
