import React from 'react';

import {PipelineJobsQueryArgs} from '@graphqlTypes';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobList from './hooks/useJobList';

type JobListProps = {
  expandActions?: boolean;
  showStatusFilter?: boolean;
} & PipelineJobsQueryArgs;

const JobList: React.FC<JobListProps> = ({
  projectId,
  pipelineId,
  expandActions = false,
  showStatusFilter = false,
}) => {
  const {loading, pipelineJobs, selectedFilters, filteredPipelineJobs} =
    useJobList({
      projectId,
      pipelineId,
    });

  return (
    <>
      {showStatusFilter && pipelineJobs.length !== 0 && !loading && (
        <JobListStatusFilter
          pipelineJobs={pipelineJobs}
          selectedFilters={selectedFilters}
        />
      )}

      <JobListStatic
        projectId={projectId}
        loading={loading}
        pipelineJobs={filteredPipelineJobs}
        expandActions={expandActions}
        listScroll
      />
    </>
  );
};

export default JobList;
