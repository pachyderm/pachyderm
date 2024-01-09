import capitalize from 'lodash/capitalize';
import React from 'react';

import {RepoInfo} from '@dash-frontend/api/pfs';
import {PipelineInfo} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import {
  usePipelineActionsMenu,
  PipelineActionsModals,
} from '@dash-frontend/components/PipelineActionsMenu';
import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import hasRepoReadPermissions from '@dash-frontend/lib/hasRepoReadPermissions';
import {readableJobState} from '@dash-frontend/lib/jobs';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Table, ButtonLink} from '@pachyderm/components';

import usePipelineListRow from './hooks/usePipelineListRow';

type PipelineListRowProps = {
  pipeline?: PipelineInfo;
  pipelineRepoMap: Record<string, RepoInfo>;
};

const getPipelineStateString = (nodeState: string, state: string) => {
  if (!nodeState || !state) {
    return '';
  } else if (state === nodeState) {
    return nodeState;
  } else {
    return `${nodeState} - ${state}`;
  }
};

const PipelineListRow: React.FC<PipelineListRowProps> = ({
  pipeline,
  pipelineRepoMap,
}) => {
  const pipelineId = pipeline?.pipeline?.name || '';
  const {
    closeRerunPipelineModal,
    rerunPipelineModalIsOpen,
    deletePipelineModalOpen,
    setDeleteModalOpen,
    onDropdownMenuSelect,
    menuItems,
    onMenuOpen,
  } = usePipelineActionsMenu(pipelineId);
  const {
    projectId,
    rolesModalOpen,
    closeRolesModal,
    hasRepoEditRoles,
    openRolesModal,
    checkPermissions,
  } = usePipelineListRow(pipelineId);
  const {searchParams, toggleSearchParamsListEntry} = useUrlQueryState();

  const addSelection = (value: string) => {
    toggleSearchParamsListEntry('selectedPipelines', value);
  };

  const access = hasRepoReadPermissions(
    pipelineRepoMap[pipelineId]?.authInfo?.permissions,
  );

  return (
    <Table.Row
      data-testid="PipelineListRow__row"
      onClick={access ? () => addSelection(pipelineId) : undefined}
      isSelected={searchParams.selectedPipelines?.includes(pipelineId)}
      hasLock={!access}
      aria-disabled={!access}
      hasCheckbox={Boolean(access)}
      overflowMenuItems={menuItems}
      dropdownOnSelect={onDropdownMenuSelect}
      openOnClick={onMenuOpen}
    >
      <Table.DataCell>{pipeline?.pipeline?.name}</Table.DataCell>
      <Table.DataCell width={180}>
        {getPipelineStateString(
          capitalize(restPipelineStateToNodeState(pipeline?.state) || ''),
          readablePipelineState(pipeline?.state || ''),
        )}
      </Table.DataCell>
      <Table.DataCell width={180}>
        {getPipelineStateString(
          capitalize(restJobStateToNodeState(pipeline?.lastJobState) || ''),
          readableJobState(pipeline?.lastJobState || ''),
        )}
      </Table.DataCell>
      <Table.DataCell width={90}>v:{pipeline?.version}</Table.DataCell>
      <Table.DataCell width={210}>
        {pipeline?.details?.createdAt
          ? getStandardDateFromISOString(pipeline?.details?.createdAt)
          : '-'}
      </Table.DataCell>
      <Table.DataCell>{pipeline?.details?.description || '-'}</Table.DataCell>
      {pipelineRepoMap[pipelineId]?.authInfo?.roles && (
        <Table.DataCell>
          <ButtonLink onClick={openRolesModal}>
            {pipelineRepoMap[pipelineId]?.authInfo?.roles?.join(', ') || 'None'}
          </ButtonLink>
        </Table.DataCell>
      )}

      <PipelineActionsModals
        pipelineId={pipelineId}
        closeRerunPipelineModal={closeRerunPipelineModal}
        rerunPipelineModalIsOpen={rerunPipelineModalIsOpen}
        deletePipelineModalOpen={deletePipelineModalOpen}
        setDeleteModalOpen={setDeleteModalOpen}
      />

      {pipelineRepoMap[pipelineId]?.authInfo?.roles && rolesModalOpen && (
        <RepoRolesModal
          show={rolesModalOpen}
          onHide={closeRolesModal}
          projectName={projectId}
          repoName={pipeline?.pipeline?.name || ''}
          readOnly={!hasRepoEditRoles}
          checkPermissions={checkPermissions}
        />
      )}
    </Table.Row>
  );
};

export default PipelineListRow;
