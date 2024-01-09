import React from 'react';

import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import IconBadge from '@dash-frontend/components/IconBadge';
import {TableViewWrapper} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {
  getStandardDateFromISOString,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import {getJobRuntime} from '@dash-frontend/lib/jobs';
import {InternalJobSet, NodeState} from '@dash-frontend/lib/types';
import {
  Table,
  StatusDotsSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
} from '@pachyderm/components';

import useRunsList from './hooks/useRunsList';
import styles from './RunsList.module.css';

type RunsListProps = {
  error?: string;
  totalJobsetsLength: number;
  jobSets?: InternalJobSet[];
};

const getJobStateBadges = (jobSet: InternalJobSet) => {
  const nodeStates = jobSet.jobs.reduce(
    (acc, job) => {
      acc[restJobStateToNodeState(job.state)] += 1;
      return acc;
    },
    {
      [NodeState.ERROR]: 0,
      [NodeState.BUSY]: 0,
      [NodeState.IDLE]: 0,
      [NodeState.PAUSED]: 0,
      [NodeState.RUNNING]: 0,
      [NodeState.SUCCESS]: 0,
    },
  );
  return (
    <span className={styles.jobStates}>
      {nodeStates[NodeState.RUNNING] > 0 && (
        <IconBadge color="green" IconSVG={StatusDotsSVG} tooltip="Running jobs">
          {nodeStates[NodeState.RUNNING]}
        </IconBadge>
      )}
      {nodeStates[NodeState.SUCCESS] > 0 && (
        <IconBadge
          color="green"
          IconSVG={StatusCheckmarkSVG}
          tooltip="Successful jobs"
        >
          {nodeStates[NodeState.SUCCESS]}
        </IconBadge>
      )}
      {nodeStates[NodeState.ERROR] > 0 && (
        <IconBadge color="red" IconSVG={StatusWarningSVG} tooltip="Failed jobs">
          {nodeStates[NodeState.ERROR]}
        </IconBadge>
      )}
    </span>
  );
};

const RunsList: React.FC<RunsListProps> = ({
  totalJobsetsLength,
  jobSets = [],
  error,
}) => {
  const {iconItems, onOverflowMenuSelect} = useRunsList();
  const {searchParams, toggleSearchParamsListEntry} = useUrlQueryState();

  const addSelection = (value: string) => {
    toggleSearchParamsListEntry('selectedJobs', value);
  };

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the jobs list"
        message="Your jobs exist, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  if (totalJobsetsLength === 0) {
    return (
      <EmptyState
        title="This DAG has no jobs"
        message="This is normal for new DAGs, but we still wanted to notify you because Pachyderm didn't detect a job on our end."
        linkToDocs={{
          text: 'Try creating a branch and pushing a commit, this automatically creates a job.',
          pathWithoutDomain: 'reference/pachctl/pachctl_create_branch/',
        }}
      />
    );
  }

  if (jobSets?.length === 0) {
    return (
      <EmptyState
        title="No matching results"
        message="We couldn't find any results matching your filters. Try using a different set of filters."
      />
    );
  }

  return (
    <TableViewWrapper>
      <Table>
        <Table.Head sticky>
          <Table.Row>
            <Table.HeaderCell>Job</Table.HeaderCell>
            <Table.HeaderCell>Job Progress</Table.HeaderCell>
            <Table.HeaderCell>Runtime</Table.HeaderCell>
            <Table.HeaderCell>ID</Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Head>

        <Table.Body aria-label="runs list">
          {jobSets.map((jobSet) => (
            <Table.Row
              key={jobSet?.job?.id}
              data-testid="RunsList__row"
              onClick={() => addSelection(jobSet?.job?.id || '')}
              isSelected={searchParams.selectedJobs?.includes(
                jobSet?.job?.id || '',
              )}
              hasCheckbox
              overflowMenuItems={iconItems}
              dropdownOnSelect={onOverflowMenuSelect(jobSet?.job?.id || '')}
            >
              <Table.DataCell width={210}>
                {jobSet?.created
                  ? getStandardDateFromISOString(jobSet?.created)
                  : '-'}
              </Table.DataCell>
              <Table.DataCell width={320}>
                {`${jobSet?.jobs.length} Job${
                  (jobSet?.jobs.length || 0) > 1 ? 's' : ''
                } Total`}
                {getJobStateBadges(jobSet)}
              </Table.DataCell>
              <Table.DataCell width={90}>
                {getJobRuntime(
                  getUnixSecondsFromISOString(jobSet.started),
                  getUnixSecondsFromISOString(jobSet.finished),
                )}
              </Table.DataCell>
              <Table.DataCell>{jobSet?.job?.id}</Table.DataCell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    </TableViewWrapper>
  );
};

export default RunsList;
