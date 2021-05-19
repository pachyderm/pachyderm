import React from 'react';

import {GetJobsQueryVariables} from '@graphqlTypes';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobList from './hooks/useJobList';

type JobListProps = {
  expandActions?: boolean;
  showStatusFilter?: boolean;
} & GetJobsQueryVariables['args'];

const JobList: React.FC<JobListProps> = ({
  projectId,
  pipelineId,
  expandActions = false,
  showStatusFilter = false,
}) => {
  const {loading, jobs, selectedFilters, filteredJobs} = useJobList({
    projectId,
    pipelineId,
  });

  return (
    <>
      {showStatusFilter && jobs.length !== 0 && !loading && (
        <JobListStatusFilter jobs={jobs} selectedFilters={selectedFilters} />
      )}

      <JobListStatic
        projectId={projectId}
        loading={loading}
        jobs={filteredJobs}
        expandActions={expandActions}
        listScroll
      />
    </>
  );
};

export default JobList;
