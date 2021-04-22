import {Form} from '@pachyderm/components';
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
  const {loading, formCtx, jobs, filteredJobs} = useJobList({
    projectId,
    pipelineId,
  });

  return (
    <Form formContext={formCtx}>
      {showStatusFilter && jobs.length !== 0 && !loading && (
        <JobListStatusFilter jobs={jobs} />
      )}

      <JobListStatic
        projectId={projectId}
        loading={loading}
        jobs={filteredJobs}
        expandActions={expandActions}
      />
    </Form>
  );
};

export default JobList;
