import {ReposWithCommitQuery} from '@graphqlTypes';
import React from 'react';

import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {Table} from '@pachyderm/components';

import useRepoListRow from './hooks/useRepoListRow';

const NO_ACCESS_TOOLTIP =
  "You currently don't have permission to view this repo. To change your permission level, contact your admin.";

type RepoListRowProps = {
  repo: ReposWithCommitQuery['repos'][number];
};

const RepoListRow: React.FC<RepoListRowProps> = ({repo}) => {
  const {
    projectId,
    generateIconItems,
    onOverflowMenuSelect,
    rolesModalOpen,
    closeRolesModal,
    editRolesPermission,
    checkRolesPermission,
  } = useRepoListRow(repo?.id || '');
  const {searchParams, updateSearchParamsAndGo} = useUrlQueryState();

  const addSelection = (value: string) => {
    if (searchParams.selectedRepos && searchParams.selectedRepos[0] === value) {
      updateSearchParamsAndGo({
        selectedRepos: [],
      });
    } else {
      updateSearchParamsAndGo({
        selectedRepos: [value],
      });
    }
  };

  const getLastCommitTimestamp = (
    repo: ReposWithCommitQuery['repos'][number],
  ) => {
    if (!repo || !repo.lastCommit) return 'N/A';

    const lastCommitDate =
      repo?.lastCommit?.finished && repo?.lastCommit?.finished > 0
        ? getStandardDate(repo?.lastCommit?.finished)
        : null;
    const lastCommitId = `${repo.lastCommit.id.slice(0, 6)}...`;

    let formatString = '';
    if (lastCommitDate) formatString += `${lastCommitDate}; `;
    formatString += lastCommitId;

    return formatString;
  };

  return (
    <Table.Row
      data-testid="RepoListRow__row"
      onClick={repo?.access ? () => addSelection(repo?.id || '') : undefined}
      isSelected={searchParams.selectedRepos?.includes(repo?.id || '')}
      hasRadio={repo?.access}
      hasLock={!repo?.access}
      aria-disabled={!repo?.access}
      lockedTooltipText={!repo?.access ? NO_ACCESS_TOOLTIP : undefined}
      overflowMenuItems={generateIconItems(repo?.lastCommit?.id)}
      dropdownOnSelect={onOverflowMenuSelect(repo)}
      openOnClick={checkRolesPermission}
    >
      <Table.DataCell>{repo?.name}</Table.DataCell>
      <Table.DataCell width={120}>{repo?.sizeDisplay || '-'}</Table.DataCell>
      <Table.DataCell width={210}>
        {repo?.createdAt && repo?.createdAt > 0
          ? getStandardDate(repo?.createdAt)
          : '-'}
      </Table.DataCell>
      <Table.DataCell width={240}>
        {getLastCommitTimestamp(repo)}
      </Table.DataCell>
      <Table.DataCell>{repo?.description}</Table.DataCell>
      {repo?.authInfo?.rolesList && (
        <Table.DataCell>
          {repo?.authInfo?.rolesList?.join(', ') || 'None'}
        </Table.DataCell>
      )}

      {repo?.authInfo?.rolesList && rolesModalOpen && (
        <RepoRolesModal
          show={rolesModalOpen}
          onHide={closeRolesModal}
          projectName={projectId}
          repoName={repo?.id}
          readOnly={!editRolesPermission}
        />
      )}
    </Table.Row>
  );
};

export default RepoListRow;
