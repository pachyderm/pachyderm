import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import JobStateBadges from '@dash-frontend/components/JobStateBadges';
import {TableViewWrapper} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {
  calculateJobSetTotalRuntime,
  getStandardDateFromISOString,
} from '@dash-frontend/lib/dateTime';
import {InternalJobSet} from '@dash-frontend/lib/types';
import {IdText, Table} from '@pachyderm/components';

import useRunsList from './hooks/useRunsList';
import styles from './RunsList.module.css';

type RunsListProps = {
  error?: string;
  totalJobsetsLength: number;
  jobSets?: InternalJobSet[];
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
                <span className={styles.jobStates}>
                  <JobStateBadges jobSet={jobSet} />
                </span>
                {`${jobSet?.jobs.length} Job${
                  (jobSet?.jobs.length || 0) > 1 ? 's' : ''
                } Total`}
              </Table.DataCell>
              <Table.DataCell width={90}>
                {calculateJobSetTotalRuntime(jobSet)}
              </Table.DataCell>
              <Table.DataCell>
                <IdText>{jobSet?.job?.id}</IdText>
              </Table.DataCell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    </TableViewWrapper>
  );
};

export default RunsList;
