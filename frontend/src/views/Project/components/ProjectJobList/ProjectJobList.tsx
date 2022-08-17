import {LoadingDots} from '@pachyderm/components';
import React, {useMemo} from 'react';

import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import JobList from '@dash-frontend/components/JobList';
import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import ProjectSidebar from '@dash-frontend/views/Project/components/ProjectSidebar';

import styles from './ProjectJobList.module.css';

const ProjectJobList: React.FC = () => {
  const {projectId} = useUrlState();
  const {viewState} = useUrlQueryState();
  const {
    jobSets,
    loading: jobSetsLoading,
    error,
  } = useJobSets({
    projectId,
  });

  const filteredJobSets = useMemo(
    () =>
      jobSets.filter(
        (job) =>
          !viewState.globalIdFilter || job.id === viewState.globalIdFilter,
      ),
    [jobSets, viewState.globalIdFilter],
  );

  return (
    <div className={styles.wrapper}>
      {jobSetsLoading ? (
        <LoadingDots />
      ) : (
        <div className={styles.jobList}>
          <JobList
            projectId={projectId}
            jobs={filteredJobSets}
            loading={jobSetsLoading}
            error={error}
            showStatusFilter
            emptyStateTitle={LETS_START_TITLE}
            emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
            cardStyle
            selectJobByDefault
          />
        </div>
      )}
      <ProjectSidebar resizable={false} />
    </div>
  );
};

export default ProjectJobList;
