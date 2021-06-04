import React from 'react';

import {PipelineJobsQueryArgs} from '@graphqlTypes';

import ListEmptyState from '../ListEmptyState';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobList from './hooks/useJobList';

type JobListProps = {
  expandActions?: boolean;
  showStatusFilter?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
} & PipelineJobsQueryArgs;

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
  const {
    loading,
    pipelineJobs,
    selectedFilters,
    filteredPipelineJobs,
    noFiltersSelected,
  } = useJobList({
    projectId,
    pipelineId,
  });

  return (
    <>
      {showStatusFilter && !loading && (
        <JobListStatusFilter
          pipelineJobs={pipelineJobs}
          selectedFilters={selectedFilters}
        />
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
          pipelineJobs={filteredPipelineJobs}
          expandActions={expandActions}
          listScroll
        />
      )}
    </>
  );
};

export default JobList;
