import {ApolloError} from '@apollo/client';
import {JobSetsQuery, NodeState} from '@graphqlTypes';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import IconBadge from '@dash-frontend/components/IconBadge';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {getJobRuntime} from '@dash-frontend/lib/jobs';
import {
  Table,
  StatusDotsSVG,
  StatusCheckmarkSVG,
  StatusWarningSVG,
} from '@pachyderm/components';

import useRunsList from './hooks/useRunsList';
import styles from './RunsList.module.css';

type RunsListProps = {
  error?: ApolloError;
  totalJobsetsLength: number;
  jobSets?: JobSetsQuery['jobSets'];
};

const getJobStateBadges = (jobSet: JobSetsQuery['jobSets'][number]) => {
  const nodeStates = jobSet.jobs.reduce(
    (acc, job) => {
      acc[job.nodeState] += 1;
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
  const {viewState, toggleSelection} = useUrlQueryState();

  const addSelection = (value: string) => {
    toggleSelection('selectedJobs', value);
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
          link: 'https://docs.pachyderm.com/latest/reference/pachctl/pachctl_create_branch/',
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
    <Table>
      <Table.Head sticky>
        <Table.Row>
          <Table.HeaderCell>Run</Table.HeaderCell>
          <Table.HeaderCell>Run Progress</Table.HeaderCell>
          <Table.HeaderCell>Runtime</Table.HeaderCell>
          <Table.HeaderCell>ID</Table.HeaderCell>
          <Table.HeaderCell />
        </Table.Row>
      </Table.Head>

      <Table.Body>
        {jobSets.map((jobSet) => (
          <Table.Row
            key={jobSet?.id}
            data-testid="RunsList__row"
            onClick={() => addSelection(jobSet?.id || '')}
            isSelected={viewState.selectedJobs?.includes(jobSet?.id || '')}
            hasCheckbox
            overflowMenuItems={iconItems}
            dropdownOnSelect={onOverflowMenuSelect(jobSet?.id || '')}
          >
            <Table.DataCell>
              {jobSet?.createdAt
                ? format(
                    fromUnixTime(jobSet?.createdAt),
                    'MMM dd, yyyy; h:mmaaa',
                  )
                : '-'}
            </Table.DataCell>
            <Table.DataCell>
              {`${jobSet?.jobs.length} Job${
                (jobSet?.jobs.length || 0) > 1 ? 's' : ''
              } Total`}
              {getJobStateBadges(jobSet)}
            </Table.DataCell>
            <Table.DataCell>
              {getJobRuntime(jobSet.createdAt, jobSet.finishedAt)}
            </Table.DataCell>
            <Table.DataCell>{jobSet?.id}</Table.DataCell>
          </Table.Row>
        ))}
      </Table.Body>
    </Table>
  );
};

export default RunsList;
