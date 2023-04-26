import {File, FileType} from '@graphqlTypes';
import capitalize from 'lodash/capitalize';
import React from 'react';

import {
  Table,
  Icon,
  ArrowRightSVG,
  FolderSVG,
  Button,
  Group,
  BasicModal,
  Link,
} from '@pachyderm/components';

import useFileActions from '../../../hooks/useFileActions';
import useFileDelete from '../../../hooks/useFileDelete';

import styles from './FileTableRow.module.css';

type FileTableRowProps = {
  selectedFiles: string[];
  addSelection: (filePath: string) => void;
  file: File;
};

const FileTableRow: React.FC<FileTableRowProps> = ({
  file,
  selectedFiles,
  addSelection,
}) => {
  const {
    deleteModalOpen,
    openDeleteModal,
    closeModal,
    deleteFile,
    loading,
    error,
  } = useFileDelete(file);
  const {fileName, filePath, onMenuSelect, iconItems} = useFileActions(
    file,
    openDeleteModal,
    true,
  );

  return (
    <Table.Row
      data-testid="FileTableRow__row"
      overflowMenuItems={file.type === FileType.FILE ? iconItems : undefined}
      dropdownOnSelect={onMenuSelect}
      // onClick={() => addSelection(file.path)}
      // isSelected={selectedFiles.includes(file.path)}
      // hasCheckbox
    >
      <Table.DataCell>
        <Link to={filePath}>
          <Group spacing={8}>
            {file.type === FileType.DIR && (
              <Icon small>
                <FolderSVG />
              </Icon>
            )}
            {fileName}
          </Group>
        </Link>
      </Table.DataCell>
      <Table.DataCell>{capitalize(file.commitAction || '-')}</Table.DataCell>
      <Table.DataCell>{file.sizeDisplay}</Table.DataCell>
      {file.type === FileType.DIR && (
        <Table.DataCell className={styles.dirArrow}>
          <Button
            to={filePath}
            buttonType="ghost"
            IconSVG={ArrowRightSVG}
            aria-label={`navigate to ${file.path}`}
          />
        </Table.DataCell>
      )}
      {deleteModalOpen && (
        <BasicModal
          show={deleteModalOpen}
          onHide={closeModal}
          headerContent="Are you sure you want to delete this File?"
          actionable
          small
          confirmText="Delete"
          onConfirm={deleteFile}
          loading={loading}
          errorMessage={error?.message}
        >
          {file.path}
          <br />
          {`${file.repoName}@${file.commitId}`}
        </BasicModal>
      )}
    </Table.Row>
  );
};

export default FileTableRow;
