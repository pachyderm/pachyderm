import {Pipeline, ReposWithCommitQuery} from '@graphqlTypes';
import capitalize from 'lodash/capitalize';
import React from 'react';

import RepoRolesModal from '@dash-frontend/components/RepoRolesModal';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {readableJobState} from '@dash-frontend/lib/jobs';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Table} from '@pachyderm/components';

import usePipelineListRow from './hooks/usePipelineListRow';

type PipelineListRowProps = {
  pipeline?: Pipeline;
  pipelineRepoMap: Record<string, ReposWithCommitQuery['repos'][0]>;
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
  const {
    projectId,
    iconItems,
    onOverflowMenuSelect,
    rolesModalOpen,
    closeRolesModal,
    editRolesPermission,
    checkRolesPermission,
  } = usePipelineListRow(pipeline?.name || '');
  const {searchParams, toggleSearchParamsListEntry} = useUrlQueryState();

  const addSelection = (value: string) => {
    toggleSearchParamsListEntry('selectedPipelines', value);
  };

  return (
    <Table.Row
      data-testid="PipelineListRow__row"
      onClick={
        pipelineRepoMap[pipeline?.id || '']?.access
          ? () => addSelection(pipeline?.name || '')
          : undefined
      }
      isSelected={searchParams.selectedPipelines?.includes(
        pipeline?.name || '',
      )}
      hasLock={!pipelineRepoMap[pipeline?.id || '']?.access}
      aria-disabled={!pipelineRepoMap[pipeline?.id || '']?.access}
      hasCheckbox={Boolean(pipelineRepoMap[pipeline?.id || '']?.access)}
      overflowMenuItems={iconItems}
      dropdownOnSelect={onOverflowMenuSelect(pipeline?.name || '')}
      openOnClick={checkRolesPermission}
    >
      <Table.DataCell>{pipeline?.name}</Table.DataCell>
      <Table.DataCell>
        {getPipelineStateString(
          capitalize(pipeline?.nodeState || ''),
          readablePipelineState(pipeline?.state || ''),
        )}
      </Table.DataCell>
      <Table.DataCell>
        {getPipelineStateString(
          capitalize(pipeline?.lastJobNodeState || ''),
          readableJobState(pipeline?.lastJobState || ''),
        )}
      </Table.DataCell>
      <Table.DataCell>v:{pipeline?.version}</Table.DataCell>
      <Table.DataCell>
        {pipeline?.createdAt ? getStandardDate(pipeline?.createdAt) : '-'}
      </Table.DataCell>
      <Table.DataCell>{pipeline?.description || '-'}</Table.DataCell>
      {pipelineRepoMap[pipeline?.id || '']?.authInfo?.rolesList && (
        <Table.DataCell>
          {pipelineRepoMap[pipeline?.id || '']?.authInfo?.rolesList?.join(
            ', ',
          ) || 'None'}
        </Table.DataCell>
      )}

      {pipelineRepoMap[pipeline?.id || '']?.authInfo?.rolesList &&
        rolesModalOpen && (
          <RepoRolesModal
            show={rolesModalOpen}
            onHide={closeRolesModal}
            projectName={projectId}
            repoName={pipeline?.name || ''}
            readOnly={!editRolesPermission}
          />
        )}
    </Table.Row>
  );
};

export default PipelineListRow;
