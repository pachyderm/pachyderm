import {LoadingDots} from '@pachyderm/components';
import React from 'react';

import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import JobList from '@dash-frontend/components/JobList';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const ProjectJobList: React.FC = () => {
  const {projectId} = useUrlState();
  const {jobSets, loading: jobSetsLoading} = useJobSets({projectId});

  return (
    <>
      {jobSetsLoading ? (
        <LoadingDots />
      ) : (
        <JobList
          projectId={projectId}
          jobs={jobSets}
          loading={jobSetsLoading}
          showStatusFilter
          emptyStateTitle={LETS_START_TITLE}
          emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
          cardStyle
          selectJobByDefault
        />
      )}
    </>
  );
};

export default ProjectJobList;
