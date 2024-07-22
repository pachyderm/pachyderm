import {FileInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
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

import useFileDelete from '../../hooks/useFileDelete';
import BranchConfirmationModal from '../BranchConfirmationModal';

import FileTableRow from './FileTableRow';
import useListViewTable from './hooks/useListViewTable';
import styles from './ListViewTable.module.css';

type ListViewTableProps = {
  files: FileInfo[];
};

const SELECT_FILES_DELETE = 'Select one or more files to multi-delete files';
const SELECT_FILES_DOWNLOAD =
  'Select one or more files to multi-download files';
const OUTPUT_REPO = 'You cannot delete files in an output repo';
const NO_BRANCH =
  'You cannot delete files from a commit that is not a branch head';

const ListViewTable: React.FC<ListViewTableProps> = ({files}) => {
  const {
    repoId,
    selectedFiles,
    addSelection,
    downloadSelected,
    noFilesSelected,
  } = useListViewTable(files);

  const {
    openDeleteConfirmationModal,
    closeDeleteConfirmationModal,
    deleteConfirmationModalOpen,
    openBranchSelectionModal,
    closeBranchSelectionModal,
    branchSelectionModalOpen,
    submitBranchSelectionForm,
    commitBranches,
    hasManyBranches,
    deleteFiles,
    deleteLoading,
    deleteError,
    isOutputRepo,
    deleteDisabled: fileDeleteDisabled,
    firstBranchId,
  } = useFileDelete(files, selectedFiles);

  const deleteDisabled = fileDeleteDisabled || noFilesSelected;

  const fileDeleteAction = hasManyBranches
    ? openBranchSelectionModal
    : openDeleteConfirmationModal;

  let tooltipText = SELECT_FILES_DELETE;
  if (isOutputRepo) {
    tooltipText = OUTPUT_REPO;
  } else if (!firstBranchId) {
    tooltipText = NO_BRANCH;
  }
  return (
    <>
      <Group spacing={8} className={styles.headerButtons}>
        <Tooltip tooltipText={tooltipText} disabled={!deleteDisabled}>
          <span>
            <Button
              IconSVG={TrashSVG}
              onClick={fileDeleteAction}
              disabled={deleteDisabled}
              buttonType="ghost"
              color="black"
              aria-label="Delete selected items"
              className={deleteDisabled && styles.disabledButton}
            />
          </span>
        </Tooltip>

        <Tooltip
          tooltipText={SELECT_FILES_DOWNLOAD}
          disabled={!noFilesSelected}
        >
          <span>
            <Button
              IconSVG={DownloadSVG}
              onClick={downloadSelected}
              disabled={noFilesSelected}
              buttonType="ghost"
              color="black"
              aria-label="Download selected items"
              className={noFilesSelected && styles.disabledButton}
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
                key={file.file?.path}
                selectedFiles={selectedFiles}
                addSelection={addSelection}
              />
            ))}
          </Table.Body>
        </Table>
      </div>

      {branchSelectionModalOpen && (
        <BranchConfirmationModal
          commitBranches={commitBranches}
          onHide={closeBranchSelectionModal}
          onSubmit={submitBranchSelectionForm}
          loading={deleteLoading}
          deleteError={deleteError}
        >
          {selectedFiles.join(', ')}
        </BranchConfirmationModal>
      )}

      {deleteConfirmationModalOpen && (
        <BasicModal
          show={deleteConfirmationModalOpen}
          onHide={closeDeleteConfirmationModal}
          headerContent={`Are you sure you want to delete the selected items from ${repoId}@${firstBranchId}?`}
          actionable
          mode="Small"
          confirmText="Delete"
          onConfirm={deleteFiles}
          loading={deleteLoading}
          errorMessage={deleteError}
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
