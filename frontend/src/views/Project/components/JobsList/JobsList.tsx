import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewFilters,
  TableViewLoadingDots,
  TableViewWrapper,
} from '@dash-frontend/components/TableView';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  getJobStateIcon,
  getVisualJobState,
  getJobRuntime,
} from '@dash-frontend/lib/jobs';
import {Table, Form} from '@pachyderm/components';

import useJobsList from './hooks/useJobsList';
import useJobsListFilters, {jobsFilters} from './hooks/useJobsListFilters';

type JobsListProps = {
  selectedJobSets?: string[];
  selectedPipelines?: string[];
  filtersExpanded: boolean;
};

const JobsList: React.FC<JobsListProps> = ({
  selectedJobSets,
  selectedPipelines,
  filtersExpanded,
}) => {
  const {projectId} = useUrlState();
  const {jobs, loading, error} = useJobs({
    projectId,
    jobSetIds: selectedJobSets,
    pipelineIds: selectedPipelines,
    details: false,
  });
  const {
    sortedJobs,
    formCtx,
    staticFilterKeys,
    clearableFiltersMap,
    multiselectFilters,
  } = useJobsListFilters({jobs});
  const {iconItems, onOverflowMenuSelect, getDatumStateBadges} = useJobsList();

  if (loading) {
    return <TableViewLoadingDots data-testid="JobsList__loadingDots" />;
  }

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the jobs list"
        message="Your jobs have been processed, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        filters={jobsFilters}
        multiselectFilters={multiselectFilters}
        clearableFiltersMap={clearableFiltersMap}
        staticFilterKeys={staticFilterKeys}
      />
      {sortedJobs?.length === 0 ? (
        <EmptyState
          title="No matching results"
          message="We couldn't find any results matching your filters. Try using a different set of filters."
        />
      ) : (
        <TableViewWrapper>
          <Table>
            <Table.Head sticky>
              <Table.Row>
                <Table.HeaderCell>ID</Table.HeaderCell>
                <Table.HeaderCell>Pipeline</Table.HeaderCell>
                <Table.HeaderCell>Datums Processed</Table.HeaderCell>
                <Table.HeaderCell>Started</Table.HeaderCell>
                <Table.HeaderCell>Duration</Table.HeaderCell>
                <Table.HeaderCell>D/L</Table.HeaderCell>
                <Table.HeaderCell>U/L</Table.HeaderCell>
                <Table.HeaderCell>Restarts</Table.HeaderCell>
                <Table.HeaderCell />
              </Table.Row>
            </Table.Head>

            <Table.Body>
              {sortedJobs.map((job) => (
                <Table.Row
                  key={`${job.id}-${job.pipelineName}`}
                  data-testid="JobsList__row"
                  overflowMenuItems={iconItems}
                  dropdownOnSelect={onOverflowMenuSelect(
                    job.id,
                    job.pipelineName,
                  )}
                >
                  <Table.DataCell width={120}>
                    {job?.id.slice(0, 6)}...
                  </Table.DataCell>
                  <Table.DataCell>
                    {getJobStateIcon(getVisualJobState(job.state), true)}
                    {' @'}
                    {job.pipelineName}
                    {' v:'}
                    {job.pipelineVersion}
                  </Table.DataCell>
                  <Table.DataCell width={320}>
                    {`${job.dataTotal} Total`}
                    {getDatumStateBadges(job)}
                  </Table.DataCell>
                  <Table.DataCell width={210}>
                    {job?.startedAt ? getStandardDate(job?.startedAt) : '-'}
                  </Table.DataCell>
                  <Table.DataCell width={120}>
                    {getJobRuntime(job.startedAt, job.finishedAt)}
                  </Table.DataCell>
                  <Table.DataCell width={120}>
                    {job.downloadBytesDisplay}
                  </Table.DataCell>
                  <Table.DataCell width={120}>
                    {job.uploadBytesDisplay}
                  </Table.DataCell>
                  <Table.DataCell width={90}>{job.restarts}</Table.DataCell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </TableViewWrapper>
      )}
    </Form>
  );
};

export default JobsList;
