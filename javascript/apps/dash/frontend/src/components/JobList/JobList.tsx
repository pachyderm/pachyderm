import React from 'react';

import {JobsQueryArgs} from '@graphqlTypes';

import ListEmptyState from '../ListEmptyState';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobList from './hooks/useJobList';

type JobListProps = {
  expandActions?: boolean;
  showStatusFilter?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
} & JobsQueryArgs;

const noJobFiltersTitle = 'Select Job Filters Above :)';
const noJobFiltersMessage =
  'No job filters are currently selected. To see any jobs, please select the available filters above.';

const JobList: React.FC<JobListProps> = ({
  projectId,
  pipelineId,
  expandActions = false,
  showStatusFilter = false,
  emptyStateTitle,
  emptyStateMessage,
}) => {
  const {loading, jobs, selectedFilters, filteredJobs, noFiltersSelected} =
    useJobList({
      projectId,
      pipelineId,
    });

  return (
    <>
      {showStatusFilter && (
        <JobListStatusFilter jobs={jobs} selectedFilters={selectedFilters} />
      )}

      {!loading && noFiltersSelected && (
        <ListEmptyState
          title={noJobFiltersTitle}
          message={noJobFiltersMessage}
        />
      )}

      {!noFiltersSelected && (
        <JobListStatic
          emptyStateTitle={emptyStateTitle}
          emptyStateMessage={emptyStateMessage}
          projectId={projectId}
          loading={loading}
          jobs={filteredJobs}
          expandActions={expandActions}
          listScroll
        />
      )}
    </>
  );
};

export default JobList;
