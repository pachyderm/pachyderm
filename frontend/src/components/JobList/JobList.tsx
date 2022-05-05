import {JobOverviewFragment, JobSetFieldsFragment} from '@graphqlTypes';
import classnames from 'classnames';
import React, {useEffect} from 'react';
import {Helmet} from 'react-helmet';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {jobRoute} from '@dash-frontend/views/Project/utils/routes';

import EmptyState from '../EmptyState';

import JobListStatic from './components/JobListStatic';
import JobListStatusFilter from './components/JobListStatusFilter';
import useJobFilters from './hooks/useJobFilters';
import styles from './JobList.module.css';

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
  cardStyle?: boolean;
  selectJobByDefault?: boolean;
}

const JobList: React.FC<JobListProps> = ({
  showStatusFilter,
  jobs,
  loading,
  emptyStateTitle,
  emptyStateMessage,
  expandActions,
  projectId,
  cardStyle,
  selectJobByDefault,
}) => {
  const {filteredJobs, selectedFilters, noFiltersSelected} = useJobFilters({
    jobs,
  });
  const browserHistory = useHistory();
  const {jobId} = useUrlState();

  useEffect(() => {
    if (
      selectJobByDefault &&
      ((!jobId && jobs.length > 0) || !jobs.find((job) => job.id === jobId))
    ) {
      jobs.length > 0 &&
        browserHistory.push(
          jobRoute({
            projectId,
            jobId: jobs[0].id,
          }),
        );
    }
  }, [browserHistory, jobId, jobs, projectId, selectJobByDefault]);

  return (
    <div className={classnames(styles.base, {[styles.cardStyle]: cardStyle})}>
      <Helmet>
        <title>Jobs - Pachyderm Console</title>
      </Helmet>

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
          cardStyle={cardStyle}
        />
      )}
    </div>
  );
};

export default JobList;
