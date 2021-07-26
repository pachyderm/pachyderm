import React from 'react';

import {JobOverviewFragment, JobSetFieldsFragment} from '@graphqlTypes';

import EmptyState from '../EmptyState';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobFilters from './hooks/useJobFilters';

const noJobFiltersTitle = 'Select Job Filters Above :)';
const noJobFiltersMessage =
  'No job filters are currently selected. To see any jobs, please select the available filters above.';

export interface JobListProps {
  expandActions?: boolean;
  showStatusFilter?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
  jobs: (JobOverviewFragment | JobSetFieldsFragment)[];
  loading: boolean;
  projectId: string;
}

const JobList: React.FC<JobListProps> = ({
  showStatusFilter,
  jobs,
  loading,
  emptyStateTitle,
  emptyStateMessage,
  expandActions,
  projectId,
}) => {
  const {filteredJobs, selectedFilters, noFiltersSelected} = useJobFilters({
    jobs,
  });

  return (
    <>
      {showStatusFilter && (
        <JobListStatusFilter jobs={jobs} selectedFilters={selectedFilters} />
      )}

      {!loading && noFiltersSelected && (
        <EmptyState title={noJobFiltersTitle} message={noJobFiltersMessage} />
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
