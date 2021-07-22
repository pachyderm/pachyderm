import React from 'react';

import {useJobs} from '@dash-frontend/hooks/useJobs';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import {
  JobOverviewFragment,
  JobSetFieldsFragment,
  JobSetsQueryArgs,
  JobsQueryArgs,
} from '@graphqlTypes';

import EmptyState from '../EmptyState';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobFilters, {JobFilters} from './hooks/useJobFilters';

const noJobFiltersTitle = 'Select Job Filters Above :)';
const noJobFiltersMessage =
  'No job filters are currently selected. To see any jobs, please select the available filters above.';

type ListProps = {
  expandActions?: boolean;
  showStatusFilter?: boolean;
  emptyStateTitle: string;
  emptyStateMessage: string;
};

type ListContent = {
  jobs: (JobOverviewFragment | JobSetFieldsFragment)[];
  filteredJobs: (JobOverviewFragment | JobSetFieldsFragment)[];
  selectedFilters: JobFilters;
  loading: boolean;
  noFiltersSelected: boolean;
  projectId: string;
} & ListProps;

const ListContent: React.FC<ListContent> = ({
  showStatusFilter,
  jobs,
  selectedFilters,
  loading,
  noFiltersSelected,
  emptyStateTitle,
  emptyStateMessage,
  filteredJobs,
  expandActions,
  projectId,
}) => {
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

type JobListProps = ListProps & JobsQueryArgs;

const JobList: React.FC<JobListProps> = ({
  projectId,
  pipelineId,
  expandActions = false,
  showStatusFilter = false,
  emptyStateTitle,
  emptyStateMessage,
}) => {
  const {jobs, loading} = useJobs({projectId, pipelineId});
  const {filteredJobs, selectedFilters, noFiltersSelected} = useJobFilters({
    jobs,
  });

  return (
    <ListContent
      projectId={projectId}
      expandActions={expandActions}
      showStatusFilter={showStatusFilter}
      emptyStateTitle={emptyStateTitle}
      emptyStateMessage={emptyStateMessage}
      jobs={jobs}
      filteredJobs={filteredJobs}
      selectedFilters={selectedFilters}
      noFiltersSelected={noFiltersSelected}
      loading={loading}
    />
  );
};

type JobSetListProps = ListProps & JobSetsQueryArgs;

export const JobSetList: React.FC<JobSetListProps> = ({
  projectId,
  expandActions = false,
  showStatusFilter = false,
  emptyStateTitle,
  emptyStateMessage,
}) => {
  const {jobSets, loading} = useJobSets({projectId});
  const {filteredJobs, selectedFilters, noFiltersSelected} = useJobFilters({
    jobs: jobSets,
  });

  return (
    <ListContent
      projectId={projectId}
      expandActions={expandActions}
      showStatusFilter={showStatusFilter}
      emptyStateTitle={emptyStateTitle}
      emptyStateMessage={emptyStateMessage}
      jobs={jobSets}
      filteredJobs={filteredJobs}
      selectedFilters={selectedFilters}
      noFiltersSelected={noFiltersSelected}
      loading={loading}
    />
  );
};

export default JobList;
