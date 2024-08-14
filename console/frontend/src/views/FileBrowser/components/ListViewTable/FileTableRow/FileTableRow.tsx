import {FileInfo, FileType} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import capitalize from 'lodash/capitalize';
import React, {useMemo} from 'react';

import {useCommitDiff} from '@dash-frontend/hooks/useCommitDiff';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {formatDiffOnlyTotals} from '@dash-frontend/lib/formatDiff';
import {
  Table,
  Icon,
  FolderSVG,
  Group,
  BasicModal,
  Link,
  SkeletonBodyText,
} from '@pachyderm/components';

import useFileActions from '../../../hooks/useFileActions';
import useFileDelete from '../../../hooks/useFileDelete';
import BranchConfirmationModal from '../../BranchConfirmationModal';

type FileTableRowProps = {
  selectedFiles: string[];
  addSelection: (filePath: string) => void;
  file: FileInfo;
};

const FileTableRow: React.FC<FileTableRowProps> = ({
  file,
  selectedFiles,
  addSelection,
}) => {
  const {repoId, projectId} = useUrlState();

  const {
    openDeleteConfirmationModal,
    closeDeleteConfirmationModal,
    deleteConfirmationModalOpen,
    openBranchSelectionModal,
    closeBranchSelectionModal,
    branchSelectionModalOpen,
    deleteFiles,
    submitBranchSelectionForm,
    deleteLoading,
    deleteError,
    hasManyBranches,
  } = useFileDelete([file]);

  const fileDeleteAction = hasManyBranches
    ? openBranchSelectionModal
    : openDeleteConfirmationModal;

  const {fileName, filePath, onMenuSelect, iconItems, commitBranches} =
    useFileActions(file, fileDeleteAction, file.fileType === FileType.FILE);

  // Using the commitId from the file object because users can navigate
  // here using the /latest path which will not have a commitId in the url.
  const {fileDiff, loading: diffLoading} = useCommitDiff(
    {
      newFile: {
        commit: {
          id: file.file?.commit?.id,
          repo: {
            name: repoId,
            project: {
              name: projectId,
            },
            type: 'user',
          },
        },
        path: file.file?.path,
      },
    },
    !!file.file?.commit?.id,
  );

  const commitAction = useMemo(() => {
    const formattedDiff = formatDiffOnlyTotals(fileDiff || []);
    const action = formattedDiff[file.file?.path || ''];
    return action ? capitalize(action) : '-';
  }, [file.file?.path, fileDiff]);

  return (
    <Table.Row
      data-testid="FileTableRow__row"
      overflowMenuItems={iconItems}
      dropdownOnSelect={onMenuSelect}
      onClick={() => addSelection(file.file?.path || '')}
      isSelected={selectedFiles.includes(file.file?.path || '')}
      hasCheckbox
    >
      <Table.DataCell>
        <Link to={filePath}>
          <Group spacing={8}>
            {file.fileType === FileType.DIR && (
              <Icon small>
                <FolderSVG />
              </Icon>
            )}
            {fileName}
          </Group>
        </Link>
      </Table.DataCell>

      <Table.DataCell>
        {diffLoading ? <SkeletonBodyText /> : commitAction}
      </Table.DataCell>
      <Table.DataCell>{formatBytes(file.sizeBytes || 0)}</Table.DataCell>

      {branchSelectionModalOpen && (
        <BranchConfirmationModal
          commitBranches={commitBranches}
          onHide={closeBranchSelectionModal}
          onSubmit={submitBranchSelectionForm}
          loading={deleteLoading}
          deleteError={deleteError}
        >
          {file.file?.path}
        </BranchConfirmationModal>
      )}

      {deleteConfirmationModalOpen && (
        <BasicModal
          show={deleteConfirmationModalOpen}
          onHide={closeDeleteConfirmationModal}
          headerContent="Are you sure you want to delete this File?"
          actionable
          mode="Small"
          confirmText="Delete"
          onConfirm={deleteFiles}
          loading={deleteLoading}
          errorMessage={deleteError}
        >
          {file.file?.path}
          <br />
          {`${file.file?.commit?.repo?.name}@${file.file?.commit?.id}`}
        </BasicModal>
      )}
    </Table.Row>
  );
};

export default FileTableRow;
