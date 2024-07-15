import React from 'react';

import {CommitInfo, RepoInfo} from '@dash-frontend/api/pfs';
import {
  useRepoActionsMenu,
  RepoActionsModals,
} from '@dash-frontend/components/RepoActionsMenu';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {
  getStandardDateFromUnixSeconds,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import formatBytes from '@dash-frontend/lib/formatBytes';
import hasRepoReadPermissions from '@dash-frontend/lib/hasRepoReadPermissions';
import {SkeletonBodyText, Table, ButtonLink} from '@pachyderm/components';

import useRepoListRow from './hooks/useRepoListRow';

const NO_ACCESS_TOOLTIP =
  "You currently don't have permission to view this repo. To change your permission level, contact your admin.";

type RepoListRowProps = {
  repo: RepoInfo;
};

const RepoListRow: React.FC<RepoListRowProps> = ({repo}) => {
  const repoId = repo?.repo?.name || '';
  const {
    deleteRepoModalOpen,
    setDeleteModalOpen,
    onDropdownMenuSelect,
    menuItems,
    onMenuOpen,
  } = useRepoActionsMenu(repoId);
  const {
    projectId,
    rolesModalOpen,
    openRolesModal,
    closeRolesModal,
    hasRepoEditRoles,
    lastCommit,
    lastCommitLoading,
    checkPermissions,
  } = useRepoListRow(repo);
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

  const getLastCommitTimestamp = (commit?: CommitInfo) => {
    if (!commit) return 'N/A';

    const lastCommitDate =
      commit?.finished &&
      getStandardDateFromUnixSeconds(
        getUnixSecondsFromISOString(commit?.finished),
      );
    const lastCommitId = `${commit?.commit?.id?.slice(0, 6)}...`;

    let formatString = '';
    if (lastCommitDate) formatString += `${lastCommitDate}; `;
    formatString += lastCommitId;

    return formatString;
  };

  const access = hasRepoReadPermissions(repo?.authInfo?.permissions);

  return (
    <Table.Row
      data-testid="RepoListRow__row"
      onClick={access ? () => addSelection(repoId) : undefined}
      isSelected={searchParams.selectedRepos?.includes(repoId)}
      hasRadio={access}
      hasLock={!access}
      aria-disabled={!access}
      lockedTooltipText={!access ? NO_ACCESS_TOOLTIP : undefined}
      overflowMenuItems={menuItems}
      dropdownOnSelect={onDropdownMenuSelect}
      openOnClick={onMenuOpen}
    >
      <Table.DataCell>{repo?.repo?.name}</Table.DataCell>
      <Table.DataCell width={120}>
        {formatBytes(repo?.details?.sizeBytes ?? repo?.sizeBytesUpperBound)}
      </Table.DataCell>
      <Table.DataCell width={120}>
        {repo?.created
          ? getStandardDateFromUnixSeconds(
              getUnixSecondsFromISOString(repo?.created),
            )
          : '-'}
      </Table.DataCell>
      <Table.DataCell width={240}>
        {lastCommitLoading ? (
          <SkeletonBodyText />
        ) : (
          getLastCommitTimestamp(lastCommit)
        )}
      </Table.DataCell>
      <Table.DataCell>{repo?.description}</Table.DataCell>
      {repo?.authInfo?.roles && (
        <Table.DataCell>
          <ButtonLink onClick={openRolesModal}>
            {repo?.authInfo?.roles?.join(', ') || 'None'}{' '}
          </ButtonLink>
        </Table.DataCell>
      )}

      <RepoActionsModals
        repoId={repoId}
        deleteRepoModalOpen={deleteRepoModalOpen}
        setDeleteModalOpen={setDeleteModalOpen}
      />

      {repo?.authInfo?.roles && rolesModalOpen && (
        <RepoRolesModal
          show={rolesModalOpen}
          onHide={closeRolesModal}
          projectName={projectId}
          repoName={repoId}
          readOnly={!hasRepoEditRoles}
          checkPermissions={checkPermissions}
        />
      )}
    </Table.Row>
  );
};

export default RepoListRow;
